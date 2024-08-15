package dns01

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	url_tool "net/url"
	"regexp"
	"strings"
	"time"

	"github.com/nicennnnnnnlee/cert_bot/dns01/common"
)

func init() {
	common.RegisterDNS01Option(option)
}

func option(ds *common.DNS01Setting) (common.DNS01, error) {
	if ds.Type != "afraid" && ds.Type != "afraid.org" {
		return nil, nil
	}
	var dns01 Afraid
	err := json.Unmarshal(ds.Config, &dns01)
	if err != nil {
		return nil, fmt.Errorf("dns config of '%s' is not valid: %v", ds.Type, err)
	}
	return &dns01, nil
}

var (
	regAfraidAnchor = regexp.MustCompile(`<a.*?>(.*?)</a>`)
	regAfraidTd     = regexp.MustCompile(`<td.*?>(.*?)</td>`)
	regAfraidVal    = regexp.MustCompile(`value=([0-9]+)`)
)

type Afraid struct {
	DSNCookie string `json:"dns_cookie"`
	DomainId  string `json:"domain_id"`
}

func (n *Afraid) UnmarshalJSON(data []byte) error {
	type alias Afraid
	obj := (*alias)(n)
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}
	if len(obj.DSNCookie) == 0 {
		return fmt.Errorf("Afraid.DSNCookie should not be empty")
	}
	if len(obj.DomainId) == 0 {
		return fmt.Errorf("Afraid.DomainId should not be empty")
	}
	return nil
}

func (af *Afraid) DeleteTXT(identifier string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("afraid: error creating request: %v", r)
		}
	}()
	// deleteFunc := func(recordId string) error {
	// 	log.Println("TODO delete recordId:", recordId)
	// 	return nil
	// }
	err = af.iteratorTxtRecord(identifier, af.deleteTXTById)
	return
}

func (af *Afraid) SetTXT(txt string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("afraid: error creating request: %v", r)
		}
	}()
	url := "https://freedns.afraid.org/subdomain/save.php?step=2"
	data := fmt.Sprintf(`type=TXT&subdomain=_acme-challenge&domain_id=%s&address=%%22%s%%22&`,
		af.DomainId, url_tool.QueryEscape(txt),
	)
	log.Println("Save Txt record: ", data)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(data)))
	if err != nil {
		return fmt.Errorf("afraid: error creating request: %v", err)
	}
	req.Header.Set("Cookie", "dns_cookie="+af.DSNCookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("afraid: error sending request: %v", err)
	}
	defer resp.Body.Close()
	html_, err := io.ReadAll(resp.Body)
	html0 := string(html_)
	if strings.Contains(html0, "<TITLE>Problems!</TITLE>") {
		errMsg, _, err := nextContent(html0, "<li>", "</font>", 0)
		if err != nil {
			return fmt.Errorf("error occurs when saving Txt record with unknown reason")
		} else {
			errMsg = html.UnescapeString(errMsg[4 : len(errMsg)-7])
			return fmt.Errorf("error occurs when saving Txt record with reason: %s", errMsg)
		}
	}
	return nil
}

func (af *Afraid) deleteTXTById(recordId string) (err error) {
	url := "https://freedns.afraid.org/subdomain/delete2.php?submit=delete&data_id[]=" + recordId
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Cookie", "dns_cookie="+af.DSNCookie)
	c := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	return err
}

func (af *Afraid) iteratorTxtRecord(identifier string, deleteFunc func(dataId string) error) error {
	url := "https://freedns.afraid.org/subdomain/"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Cookie", "dns_cookie="+af.DSNCookie)
	c := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	html0, _ := io.ReadAll(resp.Body)

	form, _, err := nextContent(string(html0), "<form", "</form>", 0)
	if err != nil {
		return fmt.Errorf("cannnot find form in html: %v", err)
	}
	index := 0
	var tr, td string
	for {
		tr, index, err = nextContent(form, "<tr", "</tr>", index)
		if err != nil {
			break
		}
		if !strings.Contains(tr, `checkbox`) {
			continue
		}
		tdIdx := 0
		td, tdIdx, _ = nextContent(tr, "<td", "</td>", tdIdx)
		recordId := regAfraidVal.FindStringSubmatch(td)[1]

		td, tdIdx, _ = nextContent(tr, "<td", "</td>", tdIdx)
		recordName := regAfraidAnchor.FindStringSubmatch(td)[1]
		if recordName != "_acme-challenge."+identifier {
			continue
		}
		td, tdIdx, _ = nextContent(tr, "<td", "</td>", tdIdx)
		recordType := regAfraidTd.FindStringSubmatch(td)[1]
		if recordType != "TXT" {
			continue
		}
		td, _, err = nextContent(tr, "<td", "</td>", tdIdx)
		if err == nil {
			recordValue := regAfraidTd.FindStringSubmatch(td)[1]
			recordValue = html.UnescapeString(recordValue)
			log.Printf("Afraid record to delete: %s %s %s %s\n", recordId, recordName, recordType, recordValue)
			err = deleteFunc(recordId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func nextContent(template, beginLabel, endLabel string, currentIdx int) (string, int, error) {
	all, err := substr(template, currentIdx)
	if err != nil {
		return "", 0, err
	}
	beginOfContent := strings.Index(all, beginLabel)
	sub, err := substr(all, beginOfContent+len(beginLabel))
	if err != nil {
		return "", 0, err
	}
	endOfContent := beginOfContent + len(beginLabel) + strings.Index(sub, endLabel) + len(endLabel)
	finalIndex := currentIdx + endOfContent
	all, err = substr(all, beginOfContent, endOfContent)
	if err != nil {
		return "", 0, err
	}
	return all, finalIndex, nil
}

func substr(str string, idx int, idxEnd ...int) (string, error) {
	if idx < 0 {
		return "", fmt.Errorf("string index(%d) must >= 0", idx)
	}
	if idx >= len(str) {
		return "", fmt.Errorf("string index(%d) must < string len(%d)", idx, len(str))
	}
	if len(idxEnd) == 0 {
		return str[idx:], nil
	}
	if idxEnd[0] >= len(str) {
		return "", fmt.Errorf("string index(%d) must < string len(%d)", idxEnd[0], len(str))
	}
	return str[idx:idxEnd[0]], nil
}
