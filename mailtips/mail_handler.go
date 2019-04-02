package mailtips

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"

	"github.com/byblix/gopro/slack"
	"github.com/byblix/gopro/utils"

	"github.com/byblix/gopro/models"

	"github.com/sendgrid/sendgrid-go"
)

// MailReq is the received Client req for mail
type MailReq struct {
	Receivers []*models.ProfileProps `json:"receivers"`
	From      *models.ProfileProps   `json:"from"`
	Subject   string                 `json:"subject"`
	Content   string                 `json:"content"`
	StoryID   string                 `json:"storyId"`
}

// MailHandler handles mail requests
// /v1/mail/send + body
func MailHandler(w http.ResponseWriter, r *http.Request) {
	mailReq := MailReq{}
	var wg sync.WaitGroup
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API"))
	err := json.NewDecoder(r.Body).Decode(&mailReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	SendMail(client, &mailReq, &wg)
	msg := createSlackMsg(&mailReq)
	err = slack.NewTipNotification(msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func createSlackMsg(mailReq *MailReq) *slack.TipSlackMsg {
	return &slack.TipSlackMsg{
		Text: "A new pro-tip has been made from: " + mailReq.From.DisplayName +
			"\nThe following medias has been tipped: " + utils.JoinStrings(unwrapMediaNames(mailReq.Receivers)),
		TitleLink: "https://app.byrd.news/" + mailReq.StoryID,
		Title:     "Story: " + mailReq.Subject,
	}
}

func unwrapMediaNames(medias []*models.ProfileProps) []string {
	output := make([]string, len(medias))
	for idx, val := range medias {
		output[idx] = val.DisplayName
	}
	return output
}

func unwrapMediaEmail(medias []*models.ProfileProps) []string {
	output := make([]string, len(medias))
	for idx, val := range medias {
		output[idx] = val.Email
	}
	return output
}
