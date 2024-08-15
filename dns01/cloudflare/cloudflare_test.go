package dns01_test

import (
	"testing"

	dns01 "github.com/nicennnnnnnlee/cert_bot/dns01/cloudflare"
)

// go test ./dns01 -v -run TestCf
func TestCf(t *testing.T) {

	cf := &dns01.Cloudflare{
		ApiEmail: "example@gmail.com", // ApiEmail不为空时，ApiKey请使用GlobalKey
		ApiKey:   "Global key or Dedicated token",
		Domain:   "Root domain in the Cloudflare panel",
		// ZoneId:    "The zone id of the domain", // 可以不填
	}
	{
		// err := cf.SetTXT("测试")
		// if err != nil {
		// 	panic(err)
		// }
	}
	{
		err := cf.DeleteTXT("test.xxx.com") // the domain is auth.Identifier.Value
		if err != nil {
			panic(err)
		}
	}
}
