package slack

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
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
