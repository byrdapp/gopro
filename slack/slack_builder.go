package slack

import (
	"log"
	"os"

	"github.com/nlopes/slack"
)

var (
	logger *log.Logger
)

// TipSlackMsg msg builder for slack msgs
type TipSlackMsg struct {
	TitleLink, Title, Text, Footer, Pretext string
}

// Colors for slack msg
var colors = map[string]string{
	"success": "#36a64f",
	"fail":    "#5f1213",
}

// NewTipNotification notifies byrd when a pro guy has tipped
func NewTipNotification(s *TipSlackMsg) error {
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
