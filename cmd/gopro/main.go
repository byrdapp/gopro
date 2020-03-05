package main

import (
	"flag"
	"os"

	"github.com/joho/godotenv"

	"github.com/blixenkrone/byrd-pro-api/internal/server"
	"github.com/blixenkrone/byrd-pro-api/pkg/logger"
)

var (
	local      = flag.Bool("local", false, "Do you want to run go run *.go with .env local file?")
	production = flag.Bool("production", false, "Is it production")
	mute       = flag.Bool("mute", false, "Mute public notificatons but not logging")
	log        = logger.NewLogger()
)

func init() {
	flag.Parse()

	if *local && !*production {
		if err := godotenv.Load(); err != nil {
			panic(err)
		}
		if *mute {
			os.Setenv("PANIC_NOTIFICATIONS", "false")
		}
		log.Infof("Running locally with %s env", os.Getenv("ENV"))
	}
}

func main() {
	s := server.NewServer()

	s.InitRoutes()

	if err := s.UseHTTP2(); err != nil {
		log.Warnf("Error with HTTP2 %s", err)
	}

	if err := s.InitDB(); err != nil {
		log.Fatalf("Error initializing DB %s", err)
	}

	s.HTTPListenServer.Addr = ":3000"
	log.Infof("Serving on host w. address %s", s.HTTPListenServer.Addr)
	// if err := s.httpListenServer.ListenAndServeTLS("./certs/insecure_cert.pem", "./certs/insecure_key.pem"); err != nil {
	if err := s.HTTPListenServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
	// * runs until os.SIGTERM happens
	s.WaitForShutdown()
}
