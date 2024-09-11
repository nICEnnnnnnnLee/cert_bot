package server

import (
	"net/http"
)

func test(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	w.Write([]byte(r.Host))
	w.Write([]byte("\n"))
	w.Write([]byte(r.TLS.ServerName))
}
