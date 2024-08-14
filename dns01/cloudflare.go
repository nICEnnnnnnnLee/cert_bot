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
)

type Cloudflare struct {
	ApiEmail string `json:"api-email"`
	ApiKey   string `json:"api-key"`
	// AccountId string `json:"accountId"`
	Domain string `json:"domain"`
	ZoneId string `json:"zoneId"`
}

type txtRecord struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
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
