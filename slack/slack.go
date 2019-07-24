package slack

import (
	"log"
	"os"

	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	models "github.com/blixenkrone/gopro/models"

	"github.com/nlopes/slack"
)

var (
	logger *log.Logger
)

// TipRequest from FE JSON req.
type TipRequest struct {
	Story      *models.StoryProps   `json:"story,omitempty"`
	Medias     []string              `json:"medias"`
	Assignment *models.Assignment   `json:"assignment"`
	Profile    *models.ProfileProps `json:"profile"`
}

// PostSlackMsg receives slack msg in body
func PostSlackMsg(w http.ResponseWriter, r *http.Request) {
	tip := &TipRequest{}
	err := json.NewDecoder(r.Body).Decode(tip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = postTip(tip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(201)
	fmt.Fprint(w, "Notified!")
}

func postTip(tip *TipRequest) error {
	slackMsg := &TipSlackMsg{
		Text: "A new pro-tip has been made from: " + tip.Profile.DisplayName +
			"\nThe following medias has been tipped: " + strings.Join(tip.Medias, ", "),
		Title:     "Story: " + tip.Story.StoryHeadline,
		TitleLink: "https://app.byrd.news/" + tip.Story.StoryID,
	}
	err := slackMsg.Success()
	if err != nil {
		return err
	}
	return nil
}

// TipSlackMsg msg builder for slack msgs
type TipSlackMsg struct {
	TitleLink, Title, Text, Footer, Pretext string
}

// Colors for slack msg
var colors = map[string]string{
	"success": "#36a64f",
	"error":   "#5f1213",
}

// Success notifies byrd when a pro guy has tipped
func (s *TipSlackMsg) Success() error {
	att := []slack.Attachment{}
	a := slack.Attachment{
		Pretext:   s.Pretext,
		Title:     s.Title,
		Color:     colors["success"],
		TitleLink: s.TitleLink,
		Fallback:  s.Text,
		Footer:    s.Footer,
	}
	att = append(att, a)
	msg := &slack.WebhookMessage{
		Text:        s.Text,
		Attachments: att,
	}

	err := slack.PostWebhook(os.Getenv("SLACK_WEBHOOK"), msg)
	if err != nil {
		return err
	}
	return nil
}
