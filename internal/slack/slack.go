package slack

import (
	"os"
	"strconv"
	"time"

	"encoding/json"
	"fmt"
	"net/http"

	"github.com/byrdapp/byrd-pro-api/public/logger"
	"github.com/nlopes/slack"
)

var (
	log = logger.NewLogger()
)

type SlackHookMsg struct {
	Msg      string
	ImageURL string
}

func Hook(msg, imgurl string) *SlackHookMsg {
	return &SlackHookMsg{
		Msg:      msg,
		ImageURL: imgurl,
	}
}

func (qs *SlackHookMsg) Panic() {
	attachment := slack.Attachment{
		Color:      "bad",
		Title:      "<!here> Pro API server paniced!",
		ImageURL:   qs.ImageURL,
		Fallback:   qs.Msg,
		AuthorName: "BUG",
		Text:       qs.Msg,
		Ts:         json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}
	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
		Channel:     "server_errors",
	}
	if err := slack.PostWebhook(os.Getenv("SLACK_WEBHOOK"), &msg); err != nil {
		log.Warn(err)
	}
}

// TipRequest from FE JSON req.
type TipRequest struct {

	// Story      *models.StoryProps   `json:"story,omitempty"`
	// Medias     []string             `json:"medias"`
	// Assignment *models.Assignment   `json:"assignment"`
	// Profile    *models.ProfileProps `json:"profile"`
}

// PostSlackMsg receives slack msg in body
func PostSlackMsg(w http.ResponseWriter, r *http.Request) {
	tip := &TipRequest{}
	err := json.NewDecoder(r.Body).Decode(tip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = postTip(tip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(201)
	fmt.Fprint(w, "Notified!")
}

func postTip(tip *TipRequest) error {
	slackMsg := &SlackMsg{
		// Text: "A new pro-tip has been made from: " + tip.DisplayName +
		// 	"\nThe following medias has been tipped: " + strings.Join(tip.Medias, ", "),
		// Title:     "Story: " + tip.StoryHeadline,
		// TitleLink: "https://app.byrd.news/" + tip.StoryID,
	}
	err := slackMsg.Success()
	if err != nil {
		return err
	}
	return nil
}

// SlackMsg msg builder for slack msgs
type SlackMsg struct {
	TitleLink, Title, Text, Footer, Pretext string
}

// Colors for slack msg
var colors = map[string]string{
	"success": "#36a64f",
	"error":   "#5f1213",
}

// Success notifies byrd when a pro guy has tipped
func (s *SlackMsg) Success() error {
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
