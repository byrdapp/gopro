package slack

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/byblix/gopro/models"
)

var (
	logger *log.Logger
)

// NotificationTip /slack/tip to slack
func NotificationTip(w http.ResponseWriter, r *http.Request) {
	// Get storyprops as JSON from CLIENT
	storyProps, err := decodeStoryProps(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalf("Fatal decoding!: %s", err)
	}
	// Create slack struct
	slackMsg := &Message{}
	// Decode struct to []struct from storyProps JSON
	slackMsg.Attachments = createAttachmentSlice(slackMsg, storyProps)
	res, err := NewSlackAttMessage(slackMsg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Fatalf("Error with slack endpoint: %s", err)
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(res.StatusCode)
	b, _ := json.Marshal(res.StatusCode)
	w.Write(b)
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
