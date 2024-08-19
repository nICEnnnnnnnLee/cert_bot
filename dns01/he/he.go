package he

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	url_tool "net/url"
	"strings"
	"time"

	"github.com/nicennnnnnnlee/cert_bot/dns01/common"
)

func init() {
	common.RegisterDNS01Option(option)
}

func option(ds *common.DNS01Setting) (common.DNS01, error) {
	if ds.Type != "he" && ds.Type != "he.net" {
		return nil, nil
	}
	var dns01 HE
	err := json.Unmarshal(ds.Config, &dns01)
	if err != nil {
		return nil, fmt.Errorf("dns config of '%s' is not valid: %v", ds.Type, err)
	}
	return &dns01, nil
}

type HE struct {
	Domain   string `json:"domain"`
	Password string `json:"password"`
}

func (n *HE) UnmarshalJSON(data []byte) error {
	type alias HE
	obj := (*alias)(n)
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}
	if len(obj.Domain) == 0 {
		return fmt.Errorf("HE.Domain should not be empty")
	}
	if len(obj.Password) == 0 {
		return fmt.Errorf("HE.Password should not be empty")
	}
	return nil
}

func (he *HE) DeleteTXT(identifier string) (err error) {
	return
}

func (he *HE) SetTXT(txt string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("he.net: error creating request: %v", r)
		}
	}()
	url := "https://dyn.dns.he.net/nic/update?hostname=_acme-challenge.%s&password=%s&txt=%s"
	url = fmt.Sprintf(url, he.Domain, he.Password, url_tool.QueryEscape(txt))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("he.net: error creating request: %v", err)
	}
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("he.net: error sending request: %v", err)
	}
	defer resp.Body.Close()
	txt_, err := io.ReadAll(resp.Body)
	txt0 := string(txt_)
	log.Println("rsp from he.net:", txt0)
	if strings.HasPrefix(txt0, "good") || strings.HasPrefix(txt0, "nochg") {
		// wait 20s, let the record cache refresh
		time.Sleep(time.Second * 10)
		log.Println("wait sometime to let the record cache refresh")
		time.Sleep(time.Second * 10)
		return nil
	} else {
		return fmt.Errorf("he.net: error set txt %s - %s", txt, txt0)
	}
}
