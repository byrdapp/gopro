package mailtips

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	sgmail "github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/byblix/gopro/models"
	"github.com/byblix/gopro/slack"
	"github.com/byblix/gopro/utils"
)

// MailReq is the received Client req for mail
type MailReq struct {
	Receivers []*models.ProfileProps `json:"receivers"`
	From      *models.ProfileProps   `json:"from"`
	Subject   string                 `json:"subject"`
	Content   string                 `json:"content"`
	StoryIDS  []string               `json:"storyIds"`
}

// MailHandler handles mail requests
// /mail/send + body
func MailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	req := MailReq{}
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API"))
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Wrong body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	resp, err := req.SendMail(client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	slack := req.createSlackMsg()
	err = slack.Success()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// MailResponse returns json for each story
type MailResponse struct {
	StoryID    string `json:"storyId"`
	StatusCode int    `json:"statusCode"`
}

// SendMail via. sendgrid
func (req *MailReq) SendMail(client *sendgrid.Client) ([]*MailResponse, error) {
	var responses = make([]*MailResponse, len(req.Receivers))
	for idx, reciever := range req.Receivers {
		from := sgmail.NewEmail(req.From.DisplayName, req.From.Email)
		subject := req.Subject
		to := sgmail.NewEmail(reciever.DisplayName, reciever.Email)
		content := req.createMailContent(reciever.Country, idx)
		htmlContent := req.createMailContent(reciever.Country, idx)
		message := sgmail.NewSingleEmail(from, subject, to, content, htmlContent)
		resp, err := client.Send(message)
		if err != nil {
			return nil, err
		}
		fmt.Println("Tipped media :" + reciever.DisplayName)
		response := &MailResponse{req.StoryIDS[idx], resp.StatusCode}
		if response != nil {
			responses = append(responses, response)
			fmt.Println(response)
		}
	}
	return responses, nil
}

func (req *MailReq) linkStoryIDS() string {
	var links = make([]string, len(req.StoryIDS))
	for i := range links {
		links = append(links, "https://app.byrd.news/story/"+req.StoryIDS[i])
	}
	return strings.Join(links, " ")
}

func (req *MailReq) createSlackMsg() *slack.TipSlackMsg {
	return &slack.TipSlackMsg{
		Text: "A new pro-tip has been made from: " + req.From.DisplayName +
			"\nThe following medias has been tipped: " + req.unwrapMediaNames(),
		TitleLink: req.linkStoryIDS(),
		Title:     "Story: " + req.linkStoryIDS(),
	}
}

func (req *MailReq) unwrapMediaNames() string {
	output := make([]string, len(req.Receivers))
	for idx, val := range req.Receivers {
		output[idx] = val.DisplayName
	}
	return utils.JoinStrings(output)
}

func (req *MailReq) createMailContent(mediaCountry string, idx int) string {
	receiverName := req.Receivers[idx].DisplayName
	mediaCountry = strings.ToLower(mediaCountry)
	countries := []string{"denmark", "sweden"}
	for i := range countries {
		if mediaCountry == countries[i] {
			mediaCountry = countries[i]
		}
	}
	switch mediaCountry {
	case "denmark":
		return fmt.Sprintf(`Hej %s, <br>
		Jeg har for nylig delt nogle stories, som kan være relevant for jer. <br>
		Klik her for at se indholdet: %s <br>
		Sig endelig til, hvis der er noget, vi kan hjælpe med på hello@byrd.news. <br>
		De bedste hilsner, <br>
		%s fra Byrd`, receiverName, req.linkStoryIDS(), req.From.DisplayName)
	case "sweden":
		return fmt.Sprintf(`Hej %s, <br>
    	Jag har precis delat innehåll som jag tror kan vara relevant för er.<br>
    	Klicka här för att se materialet: <br>
    	%s <br>
    	Säg gärna till om det är någonting vi kan hjälpa er med mail på hello@byrd.news. <br>
    	Med vänliga hälsningar, <br>
    	%s från Byrd`, receiverName, req.linkStoryIDS(), req.From.DisplayName)
	default:
		return fmt.Sprintf(`Hi %s <br>,
		I have recently shared content that might be relevant to you. <br>
    	Click to see the material: <br>
    	%s <br>
    	Please let us know if you need any help on hello@byrd.news. <br>
    	Best regards, <br>
    	%s from Byrd`, receiverName, req.linkStoryIDS(), req.From.DisplayName)
	}
}
