package dns01_test

import (
	"testing"

	dns01 "github.com/nicennnnnnnlee/cert_bot/dns01/afraid"
)

// go test ./dns01 -v -run TestAfraid
func TestAfraid(t *testing.T) {
	af := &dns01.Afraid{
		DSNCookie: "cookie from browser", // F12 find it in the browser, see https://freedns.afraid.org/faq/#17
		DomainId:  "1234567",             //	see https://freedns.afraid.org/faq/#17
	}
	{
		err := af.SetTXT("哈哈cc")
		if err != nil {
			panic(err)
		}
	}
	{
		err := af.DeleteTXT("test.xxx.com") // identifier, do not contain _acme-challenge.
		if err != nil {
			panic(err)
		}
	}
}
