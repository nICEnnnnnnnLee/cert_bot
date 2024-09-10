package server

import (
	"net/http"
	"os"
)

var UrlPrefix = GetEnvOr("UrlPrefix", "/xx")

func init() {
	http.HandleFunc(UrlPrefix+"/api/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getConfig(w, r)
			return
		}
		if r.Method == "POST" {
			setConfig(w, r)
			return
		}
	})

	http.HandleFunc(UrlPrefix+"/api/configs", getConfigs)
	http.HandleFunc(UrlPrefix+"/api/req", doCertReqDns01)
	http.HandleFunc(UrlPrefix+"/api/scripts/nginx", handeShell("nginx", "-s", "reload"))
	// http.HandleFunc(UrlPrefix+"/api/scripts/test_win", handeShell("cmd", "/c", "dir", "/b"))
	http.Handle(UrlPrefix+"/static/", handlerStaticFS())
}

func GetEnvOr(key string, defaultVal string) string {
	val, exist := os.LookupEnv(key)
	if exist {
		return val
	} else {
		return defaultVal
	}
}
