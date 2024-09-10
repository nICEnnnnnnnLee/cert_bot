package server

import (
	"io/fs"
	"net/http"

	"github.com/nicennnnnnnlee/cert_bot/public"
)

func handlerStaticFS() http.Handler {
	staticFS, _ := fs.Sub(public.FS, "static")
	return http.StripPrefix(UrlPrefix+"/static/", http.FileServer(http.FS(staticFS)))
}
