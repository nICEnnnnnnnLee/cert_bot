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

	uTest        = UrlPrefix + "/api/test"
	uConfig      = UrlPrefix + "/api/config"
	uConfigs     = UrlPrefix + "/api/configs"
	uCertReq     = UrlPrefix + "/api/req"
	uNginxReload = UrlPrefix + "/api/scripts/nginx"
	uStatic      = UrlPrefix + "/static/"
)

func init() {
	http.HandleFunc(uTest, test)
	http.HandleFunc(uConfig, handleConfig)
	http.HandleFunc(uConfigs, getConfigs)
	http.HandleFunc(uCertReq, doCertReqDns01)
	http.HandleFunc(uNginxReload, handleShell("nginx", "-s", "reload"))
	// http.HandleFunc(UrlPrefix+"/api/scripts/test_win", handleShell("cmd", "/c", "dir", "/b"))
	http.Handle(uStatic, handlerStaticFS())
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
