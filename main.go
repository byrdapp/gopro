package main

import (
	"flag"
	"os"

	storage "github.com/blixenkrone/gopro/storage"
	firebase "github.com/blixenkrone/gopro/storage/firebase"
	postgres "github.com/blixenkrone/gopro/storage/postgres"
	"github.com/blixenkrone/gopro/utils/logger"
	"github.com/joho/godotenv"
)

var (
	log = logger.NewLogger()
	pq  storage.PQService
	fb  storage.FBService

	local = flag.Bool("local", false, "Do you want to run go run *.go?")
	host  = flag.String("host", "", "What host are you using?")
	// ? not yet in use
	production = flag.Bool("production", false, "Is it production?")
)

func init() {
	// type go run *.go -local
	flag.Parse()
	if *local {
		if err := godotenv.Load(); err != nil {
			panic(err)
		}
		log.Infof("Running locally with %s variables", os.Getenv("ENV"))
	}
}

func main() {
	pqsrv, err := postgres.NewPQ()
	if err != nil {
		log.Fatalf("POSTGRESQL err: %s", err)
		return
	}
	pq = pqsrv
	defer pqsrv.Close()

	fbsrv, err := firebase.NewFB()
	if err != nil {
		log.Fatalf("Error starting firebase: %s", err)
		return
	}
	fb = fbsrv

	// Serve on localhost with localhost certs if no host provided
	s := newServer()
	if *host == "" {
		s.httpsSrv.Addr = "localhost:8085"
		log.Info("Serving on http://localhost:8085")
		// if err := s.httpsSrv.ListenAndServeTLS("./certs/insecure_cert.pem", "./certs/insecure_key.pem"); err != nil {
		if err := s.httpsSrv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}

	if err := s.useHTTP2(); err != nil {
		log.Warnf("Error with HTTP2 %s", err)
	}

	// Start a reg. HTTP on a new thread
	go func() {
		log.Info("Running http server")
		if err := s.httpSrv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Set TLS cert
	s.httpsSrv.TLSConfig.GetCertificate = s.certm.GetCertificate
	log.Info("Serving on https, authenticating for https://", *host)
	if err := s.httpsSrv.ListenAndServeTLS("", ""); err != nil {
		log.Fatal(err)
	}
}
