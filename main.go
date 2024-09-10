package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nicennnnnnnlee/cert_bot/server"
)

func main() {
	bindAddr := server.GetEnvOr("BindAddr", "127.0.0.1:8080")
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
