package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/byblix/gopro/mailtips"
	"github.com/byblix/gopro/slack"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// NewServer init routes
func NewServer() error {
	r := mux.NewRouter()
	// sendgrid mail
	r.HandleFunc("/v1/mail/send", mailtips.MailHandler).Methods("POST")
	// slack
	r.HandleFunc("/v1/slack/tip", slack.NotificationTip).Methods("POST")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprintln(w, "Nothing to see here, pal!")
	})
	fmt.Printf("Now listening to env: %s on port: %s\n", os.Getenv("ENV"), os.Getenv("PORT"))

	headersOk := handlers.AllowedHeaders([]string{"Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), handlers.CORS(headersOk, originsOk, methodsOk)(r)))
	return nil
}
