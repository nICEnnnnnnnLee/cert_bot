package common

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type DNS01Option func(*DNS01Setting) (DNS01, error)

var dns01Options = []DNS01Option{}

func RegisterDNS01Option(opts ...DNS01Option) {
	dns01Options = append(dns01Options, opts...)
}

type DNS01 interface {
	DeleteTXT(identifier string) error
	SetTXT(txt string) error
}

type DNS01Setting struct {
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

func (ds *DNS01Setting) NewDNS01() (DNS01, error) {
	for _, opt := range dns01Options {
		dns01, err := opt(ds)
		if err != nil {
			return nil, err
		}
		if dns01 == nil {
			continue
		} else {
			return dns01, nil
		}
	}
	return nil, fmt.Errorf("dns type %s is not surported", ds.Type)
}

// type NotEmptyString string

// func (nes *NotEmptyString) UnmarshalJSON(data []byte) error {
// 	str := string(data)
// 	str = strings.TrimSpace(str)
// 	if len(str) == 0 {
// 		return fmt.Errorf("string should not be blank")
// 	}
// 	*nes = NotEmptyString(str)
// 	return nil
// }

func FromFile(dns01File string) (DNS01, error) {
	log.Printf("Loading dns01 file %s", dns01File)
	raw, err := os.ReadFile(dns01File)
	if err != nil {
		return nil, fmt.Errorf("error loading dns01 file: %v", err)
	}
	return FromBytes(raw)
}

func FromBytes(raw []byte) (DNS01, error) {
	var dns01 DNS01Setting
	err := json.Unmarshal(raw, &dns01)
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
