package dns01

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/nicennnnnnnlee/cert_bot/dns01/common"
)

func init() {
	common.RegisterDNS01Option(option)
}

func option(ds *common.DNS01Setting) (common.DNS01, error) {
	if ds.Type != "cloudflare" && ds.Type != "cf" {
		return nil, nil
	}
	var dns01 Cloudflare
	log.Println("Unmarshal to Cloudflare dns01 settings")
	err := json.Unmarshal(ds.Config, &dns01)
	if err != nil {
		return nil, fmt.Errorf("dns config of '%s' is not valid: %v", ds.Type, err)
	}
	return &dns01, nil
}

type Cloudflare struct {
	ApiEmail string `json:"api-email"`
	ApiKey   string `json:"api-key"`
	Domain   string `json:"domain"`
	ZoneId   string `json:"zoneId"`
	// AccountId string `json:"accountId"`
}

func (n *Cloudflare) UnmarshalJSON(data []byte) error {
	type alias Cloudflare
	obj := (*alias)(n)
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}
	if len(obj.ApiKey) == 0 {
		return fmt.Errorf("Cloudflare.ApiKey should not be empty")
	}
	if len(obj.Domain) == 0 {
		return fmt.Errorf("Cloudflare.Domain should not be empty")
	}
	return nil
}

type txtRecord struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int `json:"ttl"`
}

func (cf *Cloudflare) setAuth(header http.Header) {
	header.Set("Content-Type", "application/json")
	if cf.ApiEmail == "" {
		log.Println("set Header: Authorization when ApiEmail is empty")
		header.Set("Authorization", "Bearer "+cf.ApiKey)
	} else {
		log.Println("set Header: X-Auth-Email/X-Auth-Key when ApiEmail is not empty")
		header.Set("X-Auth-Email", cf.ApiEmail)
		header.Set("X-Auth-Key", cf.ApiKey)
	}
}
func (cf *Cloudflare) checkConfig() error {
	if cf.Domain == "" && cf.ZoneId == "" {
		return errors.New("the Domain and ZoneId config cannot be empty at the same time")
	}
	// if cf.AccountId == "" {
	// 	log.Println("Init Cloudflare AccountId ...")
	// 	err := cf.initAccountId()
	// 	log.Println("cf.AccountId err:", err)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// log.Println("cf.AccountId:", cf.AccountId)
	if cf.ZoneId == "" {
		err := cf.initZoneId()
		if err != nil {
			return err
		}
	}
	log.Println("cf.ZoneId:", cf.ZoneId)
	return nil
}

func (cf *Cloudflare) DeleteTXT(identifier string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cloudflare: error creating request: %v", r)
		}
	}()
	if err = cf.checkConfig(); err != nil {
		return err
	}
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=_acme-challenge.%s&type=TXT",
		cf.ZoneId, identifier)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	cf.setAuth(req.Header)
	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	_ = json.Unmarshal(body, &result)
	records := result["result"].([]interface{})
	for idx := range records {
		id := records[idx].(map[string]interface{})["id"]
		log.Println("records id to delete: ", id)
		url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", cf.ZoneId, id)
		req, _ := http.NewRequest(http.MethodDelete, url, nil)
		cf.setAuth(req.Header)
		resp, _ := c.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(body, &result)
		if !result["success"].(bool) {
			return fmt.Errorf("cloudflare: error deleting txt record: %v", result)
		}
	}
	return nil
}

func (cf *Cloudflare) SetTXT(txt string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cloudflare: error creating request: %v", r)
		}
	}()
	if err = cf.checkConfig(); err != nil {
		return err
	}
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", cf.ZoneId)
	record := txtRecord{
		Type:    "TXT",
		Name:    "_acme-challenge",
		Content: txt,
		TTL:     60,
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("cloudflare: error creating request: %v", err)
	}
	cf.setAuth(req.Header)
	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("cloudflare: error sending request: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("cloudflare: error rsp StatusCode: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cloudflare: error parsing response body: %v", err)
	}
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return fmt.Errorf("cloudflare: error parsing response body: %v", err)
	}
	if result["success"].(bool) {
		return nil
	} else {
		return fmt.Errorf("cloudflare: error parsing response body: %v", result)
	}
}

// func (cf *Cloudflare) initAccountId() (err error) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			err = fmt.Errorf("cloudflare: error creating request: %v", r)
// 		}
// 	}()
// 	url := "https://api.cloudflare.com/client/v4/accounts?page=1&per_page=20&direction=desc"
// 	req, err := http.NewRequest(http.MethodGet, url, nil)
// 	if err != nil {
// 		return fmt.Errorf("cloudflare: error creating request: %v", err)
// 	}
// 	cf.setAuth(req.Header)
// 	c := &http.Client{Timeout: 10 * time.Second}
// 	resp, err := c.Do(req)
// 	if err != nil {
// 		return fmt.Errorf("cloudflare: error sending request: %v", err)
// 	}
// 	if resp.StatusCode != 200 {
// 		return fmt.Errorf("cloudflare: error rsp StatusCode: %v", resp.StatusCode)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return fmt.Errorf("cloudflare: error parsing request body: %v", err)
// 	}
// 	var result map[string]interface{}
// 	err = json.Unmarshal(body, &result)
// 	if err != nil {
// 		return fmt.Errorf("cloudflare: error parsing request body: %v", err)
// 	}
// 	cf.AccountId = result["result"].([]interface{})[0].(map[string]interface{})["id"].(string)
// 	return nil
// }

func (cf *Cloudflare) initZoneId() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cloudflare: error creating request: %v", r)
		}
	}()
	url := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/zones?name=%s&page=1&per_page=20&order=status&direction=desc&match=all",
		cf.Domain,
		// cf.AccountId, cf.Domain,
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("cloudflare: error creating request: %v", err)
	}
	cf.setAuth(req.Header)
	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("cloudflare: error sending request: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("cloudflare: error rsp StatusCode: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cloudflare: error parsing request body: %v", err)
	}
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return fmt.Errorf("cloudflare: error parsing request body: %v", err)
	}
	//  res["result"][0]["id"]
	cf.ZoneId = result["result"].([]interface{})[0].(map[string]interface{})["id"].(string)
	return nil
}
