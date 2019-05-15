package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"time"

	postgres "github.com/byblix/gopro/storage/postgres"
	mux "github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

var db postgres.Service

var (
	local = flag.Bool("local", false, "Do you want to run go run *.go?")
	host  = flag.String("host", "", "What host are you using?")
	// ? not yet in use
	production = flag.Bool("production", false, "Is it production?")
)

func init() {
	// type go run *.go -local
	flag.Parse()
	if *local {
		log.Info("Running locally")
		if err := godotenv.Load(); err != nil {
			log.Fatalln(err)
		}
	}
}

func main() {
	svc, err := postgres.NewPQ()
	if err != nil {
		log.Fatalf("POSTGRESQL err: %s", err)
	}
	db = svc
	defer svc.Close()
	s, err := newServer()
	if err != nil {
		log.Fatal(err)
	}

	// Serve on localhost with localhost certs if no host provided
	if *host == "" {
		s.httpsSrv.Addr = "localhost:8085"
		logrus.Info("Serving on http://localhost:8085")
		if err := s.httpsSrv.ListenAndServe(); err != nil {
			s.log.Fatal(err)
		}
	}

	if err := s.useHTTP2(); err != nil {
		s.log.Warnf("Error with HTTP2 %s", err)
	}

	// Start a reg. HTTP on a new thread
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil {
			s.log.Fatal(err)
		}
	}()

	// Set TLS cert
	s.httpsSrv.TLSConfig.GetCertificate = s.certm.GetCertificate
	s.log.Info("Serving on https, authenticating for https://", *host)
	if err := s.httpsSrv.ListenAndServeTLS("", ""); err != nil {
		s.log.Fatal(err)
	}

	// headersOk := handlers.AllowedHeaders([]string{"content-type"})
	// originsOk := handlers.AllowedOrigins([]string{"*"})
	// methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})
	// handler := handlers.CORS(headersOk, originsOk, methodsOk)(r)

}

func getMediaByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		params := mux.Vars(r)
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()
		val, err := db.GetMediaByID(ctx, params["id"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if err := json.NewEncoder(w).Encode(val); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func createMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.Header.Set("content-type", "application/json")
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()
		var media postgres.Media
		if err := json.NewDecoder(r.Body).Decode(&media); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		id, err := db.CreateMedia(ctx, &media)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		err = json.NewEncoder(w).Encode(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// getMedias endpoint: /medias
func getMedias(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("content-type", "application/json")
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()
	// todo: params
	medias, err := db.GetMedias(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := json.NewEncoder(w).Encode(medias); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/**
 * * What is the relationship between media and department?
 * * What should be shown to the user of these ^?
 * ! implement context in all server>db calls
 */
