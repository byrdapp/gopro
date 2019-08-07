package slack

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type (
	// Message .
	Message struct {
		Text        string        `json:"text"`
		Attachments []Attachments `json:"attachments"`
	}
	// Attachments -
	Attachments struct {
		// https://api.slack.com/docs/messages/builder
		Fallback   string    `json:"fallback"` //Required!
		Text       string    `json:"text"`     //Within attachment
		Pretext    string    `json:"pretext"`  //Outside attachment
		Color      string    `json:"color"`
		Authorname string    `json:"author_name"`
		Authoricon string    `json:"author_icon,omitempty"`
		Title      string    `json:"title"`
		Titlelink  string    `json:"title_link"`
		Fields     []*Fields `json:"fields"`
		Timestamp  int64     `json:"ts"`
	}
	// Fields .
	Fields struct {
		Title string `json:"title"`
		Value string `json:"value"`
		Short bool   `json:"short"`
	}
)

// NewSlackAttMessage happens if a professional tipped a media
func NewSlackAttMessage(i *Message) (*http.Response, error) {
	JSON, err := marshallSlackMsg(i)
	if err != nil {
		return nil, err
	}
	res, err := newRequest(JSON)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func marshallSlackMsg(i interface{}) ([]byte, error) {
	slackJSON, err := json.Marshal(i)
	if err != nil {
		log.Panicf("Error marshalling JSON: %s", err)
	}
	return slackJSON, nil
}

func newRequest(slackJSON []byte) (*http.Response, error) {
	hookURL := os.Getenv("SLACK_WEBHOOK")
	client := &http.Client{}
	req, err := http.NewRequest("POST", hookURL, bytes.NewBuffer(slackJSON))
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-type", "application/json")
	defer resp.Body.Close()
	return resp, nil
}
