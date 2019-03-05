package slack

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/byblix/gopro/utils"

	"github.com/byblix/gopro/models"
)

// TipRequest from FE JSON req.
type TipRequest struct {
	Story      *models.StoryProps   `json:"story,omitempty"`
	Medias     []string             `json:"medias"`
	Assignment *models.Assignment   `json:"assignment"`
	Profile    *models.ProfileProps `json:"profile"`
}

// HandleSlackMessage receives slack msg in body
func HandleSlackMessage(w http.ResponseWriter, r *http.Request) {
	tip := &TipRequest{}
	err := json.NewDecoder(r.Body).Decode(tip)
	if err != nil {
		fmt.Fprint(w, http.StatusInternalServerError)
	}

	err = postTip(tip)
	if err != nil {
		fmt.Fprint(w, err)
	}

	w.WriteHeader(201)
	fmt.Fprint(w, "Notified!")
}

func postTip(tip *TipRequest) error {
	slackMsg := &TipSlackMsg{
		Text: "A new pro-tip has been made from: " + tip.Profile.DisplayName +
			"\nThe following medias has been tipped: " + utils.JoinStrings(tip.Medias),
		Title:     "Story: " + tip.Story.StoryHeadline,
		TitleLink: "https://app.byrd.news/" + tip.Story.StoryID,
	}
	err := NewTipNotification(slackMsg)
	if err != nil {
		return err
	}
	return nil
}
