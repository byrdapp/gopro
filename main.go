package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/byblix/gopro/mailtips"
	"github.com/byblix/gopro/slack"
	postgres "github.com/byblix/gopro/storage/postgres"
	"golang.org/x/crypto/acme/autocert"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// to run this locally with dev: $ go build && ./gopro -env="local"

var db postgres.Service

func main() {
	if err := InitEnvironment(); err != nil {
		logrus.Fatalln(err)
	}
	svc, err := postgres.NewPQ()
	db = svc
	if err != nil {
		logrus.Fatalf("POSTGRESQL err: %s", err)
	}
	defer svc.Close()

	r := mux.NewRouter()
	r.HandleFunc("/mail/send", mailtips.MailHandler).Methods("POST")
	r.HandleFunc("/slack/tip", slack.PostSlackMsg).Methods("POST")
	r.HandleFunc("/media", getMedias).Methods("GET")
	r.HandleFunc("/media/{id}", getMediaByID).Methods("GET")
	r.HandleFunc("/media", createMedia).Methods("POST")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooEarly)
		fmt.Fprintln(w, "Nothing to see here :-)")
	})
	fmt.Printf("Now listening to env: %s on port: %s\n", os.Getenv("ENV"), os.Getenv("PORT"))

	// https://medium.com/weareservian/automagical-https-with-docker-and-go-4953fdaf83d2
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("cert-cache"),
		HostPolicy: autocert.HostWhitelist("gopro.byrd.news"),
	}

	server := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: r,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	headersOk := handlers.AllowedHeaders([]string{"content-type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})
	handler := handlers.CORS(headersOk, originsOk, methodsOk)(r)
	_ = handler
	err = server.ListenAndServe()
	if err != nil {
		logrus.Fatal(err)
	}

}

// InitEnvironment : set the cli flag -env=local if must be run locally
func InitEnvironment() error {
	env, ok := os.LookupEnv("ENV")
	flag.StringVar(&env, "env", env, "Environment used")
	flag.Parse()

	if env == "local" {
		if err := godotenv.Load(); err != nil {
			return err
		}
	} else {
		if !ok {
			return errors.New("No environment provided")
		}
		fmt.Println("Server CFG is being used")
	}

	fmt.Printf("%s environment is used as config\n", env)
	return nil
}

func getMediaByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()
	val, err := db.GetMediaByID(ctx, params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := json.NewEncoder(w).Encode(val); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createMedia(w http.ResponseWriter, r *http.Request) {
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
