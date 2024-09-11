package dns01

import (
	_ "github.com/nicennnnnnnlee/cert_bot/dns01/afraid"
	_ "github.com/nicennnnnnnlee/cert_bot/dns01/cloudflare"
	"github.com/nicennnnnnnlee/cert_bot/dns01/common"
	_ "github.com/nicennnnnnnlee/cert_bot/dns01/he"
)

var FromFile = common.FromFile

type DNS01Setting struct {
	common.DNS01Setting
}
