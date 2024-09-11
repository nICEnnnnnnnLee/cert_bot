package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

func Main() {
	bindAddr := GetEnvOr("BindAddr", "127.0.0.1:8080")
	log.Println("Running service at " + bindAddr)
	server := &http.Server{
		Addr: bindAddr,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			// panic(err)
			log.Println(err)
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signalCh
	log.Printf("Received signal: %v\n", sig)

	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server shutdown failed: %v\n", err)
	}
	log.Println("Server shutdown gracefully")

}
