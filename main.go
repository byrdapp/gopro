package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/byblix/gopro/mailtips"
	"github.com/byblix/gopro/slack"
	psql "github.com/byblix/gopro/storage/postgres"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
)

// to run this locally with dev: $ go build && ./gopro -env="local"

var db *psql.Postgres

func main() {
	if err := InitEnvironment(); err != nil {
		logrus.Fatalln(err)
	}
	db, err := psql.NewPQ()
	if err != nil {
		logrus.Fatalf("POSTGRESQL err: %s", err)
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/mail/send", mailtips.MailHandler).Methods("POST")
	r.HandleFunc("/slack/tip", slack.PostSlackMsg).Methods("POST")
	r.HandleFunc("/media/{id}", getMediaByID).Methods("GET")
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
	_ = handler
	err = server.ListenAndServe()
	if err != nil {
		logrus.Fatal(err)
	}
}

// InitEnvironment : set the cli flag -env=local if must be run locally
func InitEnvironment() error {
	env := os.Getenv("ENV")
	flag.StringVar(&env, "env", env, "Environment used")
	flag.Parse()
	if env == "local" {
		if err := godotenv.Load(); err != nil {
			return err
		}
	}
	fmt.Println(os.Getenv("ENV") + " " + "CFG is loaded")
	return nil
}

func getMediaByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	val, err := db.GetMediaByID(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Println(val)
}
