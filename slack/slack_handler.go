package slack

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/byblix/gopro/models"
	"github.com/nlopes/slack"
)

var (
	logger *log.Logger
)

// MsgBuilder msg builder for slack msgs
type MsgBuilder struct {
	TitleLink, Period, Text, Color, Footer, Pretext string
}

// NewTipNotification notifies people when theres a newly generated PDF
func NewTipNotification(s *MsgBuilder) error {
	att := []slack.Attachment{}
	a := slack.Attachment{
		Pretext:   s.Pretext,
		Title:     s.Period,
		TitleLink: s.TitleLink,
		Color:     s.Color,
		Fallback:  s.Text,
		Footer:    s.Footer,
	}
	att = append(att, a)
	msg := &slack.WebhookMessage{
		Text:        s.Text,
		Attachments: att,
	}

	if err := slack.PostWebhook(os.Getenv("SLACK_WEBHOOK"), msg); err != nil {
		return err
	}
	return nil
}

func decodeStoryProps(reader io.Reader) (*models.StoryProps, error) {
	storyProps := &models.StoryProps{}
	d := json.NewDecoder(reader)
	if err := d.Decode(storyProps); err != nil {
		return nil, err
	}
	return storyProps, nil
}

// Creating the body for message attachment
func createAttachmentSlice(msg *Message, storyProps *models.StoryProps) []Attachments {
	attMsg := &Attachments{
		Fallback:   "Some pro tipped a Media about their story!",
		Color:      Colors["success"],
		Timestamp:  time.Now().Unix(),
		Authoricon: storyProps.ProfileProps.ProfilePicture,
		Title:      "Some pro photographer just tipped a media!",
		Text:       storyProps.ProfileProps.DisplayName + " tipped the medias: " + fullMediaList(storyProps.Assignment.Medias) + " with a story!",
		Fields:     createFieldsSlice(storyProps),
	}
	return append(msg.Attachments, *attMsg)
}

func createFieldsSlice(values *models.StoryProps) []*Fields {
	var output []*Fields
	fields := &Fields{
		Title: values.Headline,
		Value: "https://app.byrd.news/" + values.StoryID,
		Short: false,
	}
	return append(output, fields)
}

func fullMediaList(list []string) string {
	return strings.Join(list, ", ")
}
