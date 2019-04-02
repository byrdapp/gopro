package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/crypto/acme/autocert"

	"github.com/byblix/gopro/mailtips"
	"github.com/byblix/gopro/slack"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// NewServer init routes
func NewServer() error {
	r := mux.NewRouter()
	r.HandleFunc("/v1/mail/send", mailtips.MailHandler).Methods("POST")
	r.HandleFunc("/v1/slack/tip", slack.PostSlackMsg).Methods("POST")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprintln(w, "Nothing to see here :-)")
	})
	fmt.Printf("Now listening to env: %s on port: %s\n", os.Getenv("ENV"), os.Getenv("PORT"))

	// https://medium.com/weareservian/automagical-https-with-docker-and-go-4953fdaf83d2
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("cert-cache"),
		HostPolicy: autocert.HostWhitelist("go-service.byrd.news"),
	}

	server := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: r,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	headersOk := handlers.AllowedHeaders([]string{"Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	handler := handlers.CORS(headersOk, originsOk, methodsOk)(r)
	if err := http.ListenAndServe(":80", certManager.HTTPHandler(handler)); err != nil {
		return err
	}
	server.ListenAndServeTLS("", "")

	return nil
}
