package main

import (
	"github.com/nicennnnnnnlee/cert_bot/cli"
	"github.com/nicennnnnnnlee/cert_bot/server"
)

func main() {
	mode := server.GetEnvOr("Mode", "server")
	if mode == "server" {
		server.Main()
	} else {
		cli.Main()
	}
}
