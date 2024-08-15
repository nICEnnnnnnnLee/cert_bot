package dns01

import (
	_ "github.com/nicennnnnnnlee/cert_bot/dns01/afraid"
	_ "github.com/nicennnnnnnlee/cert_bot/dns01/cloudflare"
	"github.com/nicennnnnnnlee/cert_bot/dns01/common"
)

var FromFile = common.FromFile
var FromBytes = common.FromBytes
