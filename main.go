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

	local = flag.Bool("local", false, "Do you want to run go run *.go with .env local file?")
	host  = flag.String("host", "pro.development.byrd.news", "What host are you using (pro.byrd.news?)?")
	// ? not yet in use
	production = flag.Bool("production", false, "Is it production?")
	ssl        = flag.Bool("ssl", false, "To set ssl or not?")
)

func init() {
	// type go run *.go -local
	flag.Parse()
	if *local && !*production {
		if err := godotenv.Load(); err != nil {
			panic(err)
		}
		log.Infof("Running locally with %s env", os.Getenv("ENV"))
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
	if !*ssl {
		s.httpListenServer.Addr = ":8080"
		log.Infof("Serving probably locally on host w. address %s", s.httpListenServer.Addr)
		// if err := s.httpListenServer.ListenAndServeTLS("./certs/insecure_cert.pem", "./certs/insecure_key.pem"); err != nil {
		if err := s.httpListenServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}

	if err := s.useHTTP2(); err != nil {
		log.Warnf("Error with HTTP2 %s", err)
	}

	// Start a reg. HTTP on a new thread
	go func() {
		log.Info("Running http server")
		if err := s.httpRedirectServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	s.httpListenServer.Addr = ":8080"
	s.httpListenServer.TLSConfig.GetCertificate = s.certm.GetCertificate
	log.Infof("Serving on a server with host: %s and address %s", *host, s.httpListenServer.Addr)
	if err := s.httpListenServer.ListenAndServeTLS("", ""); err != nil {
		log.Fatal(err)
	}
}
