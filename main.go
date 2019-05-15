package main

import (
	"encoding/json"
	"net/http"
	"time"

	postgres "github.com/byblix/gopro/storage/postgres"
	"github.com/byblix/gopro/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

var db postgres.Service

func main() {
	if err := utils.InitEnvironment(); err != nil {
		logrus.Fatalln(err)
	}
	svc, err := postgres.NewPQ()
	if err != nil {
		logrus.Fatalf("POSTGRESQL err: %s", err)
	}
	db = svc
	defer svc.Close()

	if err := newServer(); err != nil {
		logrus.Fatalf("Error starting server:%s", err)
	}

	// headersOk := handlers.AllowedHeaders([]string{"content-type"})
	// originsOk := handlers.AllowedOrigins([]string{"*"})
	// methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})
	// handler := handlers.CORS(headersOk, originsOk, methodsOk)(r)

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
