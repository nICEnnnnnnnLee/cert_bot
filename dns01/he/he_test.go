package he_test

import (
	"testing"

	dns01 "github.com/nicennnnnnnlee/cert_bot/dns01/he"
)

// go test ./dns01/he -v -run TestHE
func TestHE(t *testing.T) {
	he := &dns01.HE{
		Domain:   "domain for dns01 challenge",
		Password: "password for specific txt record", // see https://dns.he.net/docs.html
	}
	{
		err := he.SetTXT("cccc()")
		if err != nil {
			panic(err)
		}
		err = he.SetTXT("cccc()2")
		if err != nil {
			panic(err)
		}
	}
	// {
	// 	err := af.DeleteTXT("test.xxx.com") // identifier, do not contain _acme-challenge.
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }
}
