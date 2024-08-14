package dns01

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type DNS01 interface {
	DeleteTXT(identifier string) error
	SetTXT(txt string) error
}

type DNS01Setting struct {
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

func (ds *DNS01Setting) NewDNS01() (DNS01, error) {
	switch ds.Type {
	case "cf":
		fallthrough
	case "cloudflare":
		var cf Cloudflare
		err := json.Unmarshal(ds.Config, &cf)
		if err != nil {
			return nil, fmt.Errorf("dns config of '%s' is not valid: %v", ds.Type, err)
		}
		return &cf, nil
	default:
		return nil, fmt.Errorf("dns type %s is not surported", ds.Type)
	}
}

func FromFile(dns01File string) (DNS01, error) {
	log.Printf("Loading dns01 file %s", dns01File)
	raw, err := os.ReadFile(dns01File)
	if err != nil {
		return nil, fmt.Errorf("error loading dns01 file: %v", err)
	}
	var dns01 DNS01Setting
	err = json.Unmarshal(raw, &dns01)
	if err != nil {
		return nil, fmt.Errorf("no valid dns01 config json provided: %v", err)
	}
	return dns01.NewDNS01()
}

//	func getCf(confBase64Str string) (*dns01.Cloudflare, error) {
//		b64, err := base64.RawStdEncoding.DecodeString(cloudflareConfigBase64)
//		cf := &dns01.Cloudflare{}
//		if err != nil {
//			return nil, fmt.Errorf("no valid cloudflare config base64 provided")
//		}
//		log.Println(string(b64))
//		err = json.Unmarshal(b64, cf)
//		if err != nil {
//			return nil, fmt.Errorf("no valid cloudflare config json provided: %v", err)
//		}
//		log.Println(cf)
//		return cf, nil
//	}
