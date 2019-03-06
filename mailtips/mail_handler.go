package mailtips

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/byblix/gopro/models"

	"github.com/sendgrid/rest"

	"github.com/sendgrid/sendgrid-go"
)

// MailReq is the received Client req for mail
type MailReq struct {
	Receivers []*models.ProfileProps `json:"receivers"`
	From      *models.ProfileProps   `json:"from"`
	Subject   string                 `json:"subject"`
	Content   string                 `json:"content"`
}

// MailHandler handles mail requests
// /v1/mail/send + body
func MailHandler(w http.ResponseWriter, r *http.Request) {
	mailReq := MailReq{}
	var wg sync.WaitGroup
	ch := make(chan *rest.Response)
	// client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API"))
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API"))
	err := json.NewDecoder(r.Body).Decode(&mailReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = SendMail(client, &mailReq, ch, &wg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	resp := <-ch
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(resp.StatusCode)
	fmt.Fprintln(w, resp.StatusCode)
}
