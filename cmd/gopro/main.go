package main

import (
	"flag"
	"os"

	"github.com/joho/godotenv"

	"github.com/blixenkrone/gopro/internal/server"
	"github.com/blixenkrone/gopro/pkg/logger"
)

// ssl        = flag.Bool("ssl", false, "To set ssl or not?")
var (
	local      = flag.Bool("local", false, "Do you want to run go run *.go with .env local file?")
	production = flag.Bool("production", false, "Is it production?")
	startdb    = flag.Bool("startdb", true, "Start with DB connection")
	log        = logger.NewLogger()
)

func init() {
	flag.Parse()

	if *local && !*production {
		if err := godotenv.Load(); err != nil {
			panic(err)
		}
		log.Infof("Running locally with %s env", os.Getenv("ENV"))
	}
}

func main() {
	s := server.NewServer()

	if err := s.UseHTTP2(); err != nil {
		log.Warnf("Error with HTTP2 %s", err)
	}

	if *startdb {
		if err := s.InitDB(); err != nil {
			log.Fatalf("Error initializing DB %s", err)
		}
	}

	s.HttpListenServer.Addr = ":3000"
	log.Infof("Serving on host w. address %s", s.HttpListenServer.Addr)
	// if err := s.httpListenServer.ListenAndServeTLS("./certs/insecure_cert.pem", "./certs/insecure_key.pem"); err != nil {
	if err := s.HttpListenServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
	// * runs until os.SIGTERM happens
	s.WaitForShutdown()
}
