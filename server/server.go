package server

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bddjr/hlfhr"
)

var (
	UrlPrefix = GetEnvOr("UrlPrefix", "/xx")
	bindAddr  = GetEnvOr("BindAddr", "127.0.0.1:8080")
	certPath  = GetEnvOr("CertPath", "")
	keyPath   = GetEnvOr("KeyPath", "")
)

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

	http.HandleFunc(UrlPrefix+"/api/test", test)
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

func startServer(s *http.Server) error {
	if certPath != "" && keyPath != "" {
		tlsCert := &TlsCert{
			CertPath:       certPath,
			KeyPath:        keyPath,
			AttempDuration: time.Minute * 5,
		}
		s.TLSConfig = &tls.Config{
			GetCertificate: tlsCert.GetCertFunc(),
		}

		l, err := net.Listen("tcp", s.Addr)
		if err != nil {
			return err
		}
		defer l.Close()

		// Use hlfhr.NewListener
		l = hlfhr.NewListener(l, s, nil)
		return s.ServeTLS(l, "", "")
	} else {
		return s.ListenAndServe()
	}
}

type exitSig struct{}

func (exitSig) String() string {
	return "Exit signal"
}
func (exitSig) Signal() {}

func Main() {

	log.Println("Running service at " + bindAddr)
	server := &http.Server{
		Addr: bindAddr,
	}
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := startServer(server); err != nil {
			// panic(err)
			log.Println(err)
			signalCh <- exitSig{}
		}
	}()

	sig := <-signalCh
	log.Printf("Received signal: %v\n", sig)

	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server shutdown failed: %v\n", err)
	}
	log.Println("Server shutdown gracefully")

}
