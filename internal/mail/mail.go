package mail

import (
	"fmt"
	"strings"

	"github.com/blixenkrone/byrd-pro-api/pkg/conversion"

	"github.com/sendgrid/sendgrid-go"
	sgmail "github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/blixenkrone/byrd-pro-api/internal/slack"
	"github.com/blixenkrone/byrd-pro-api/internal/storage"
)

// RequestBody is the received Client req for mail
type RequestBody struct {
	Recievers []*storage.FirebaseProfile `json:"recievers"`
	From      *storage.FirebaseProfile   `json:"from"`
	Subject   string                     `json:"subject"`
	Content   string                     `json:"content"`
	StoryIDS  []string                   `json:"storyIds"`
}

// Response returns json for each story
type Response struct {
	Receiver   string `json:"receiver"`
	StatusCode int    `json:"statusCode"`
}

// SendMail via. sendgrid
func (req *RequestBody) SendMail(client *sendgrid.Client) ([]*Response, error) {
	var responses []*Response
	for idx, reciever := range req.Recievers {
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
		response := &Response{reciever.DisplayName, resp.StatusCode}
		responses = append(responses, response)
		fmt.Printf("%s tipped media %s", req.From.DisplayName, response.Receiver)
	}
	for _, v := range responses {
		fmt.Println(v)
	}
	// Create slack msg
	slack := req.createSlackMsg()
	if err := slack.Success(); err != nil {
		return nil, err
	}
	return responses, nil
}

func (req *RequestBody) linkStoryIDS() string {
	var links = make([]string, len(req.StoryIDS))
	for i := range links {
		links = append(links, "https://app.byrd.news/story/"+req.StoryIDS[i])
	}
	return strings.Join(links, " ")
}

// CreateSlackMsg -
func (req *RequestBody) createSlackMsg() *slack.SlackMsg {
	return &slack.SlackMsg{
		Text: "A new pro-tip has been made from: " + req.From.DisplayName +
			"\nThe following medias has been tipped: " + req.unwrapMediaNames(),
		TitleLink: req.linkStoryIDS(),
		Title:     "Story: " + req.linkStoryIDS(),
	}
}

func (req *RequestBody) unwrapMediaNames() string {
	output := make([]string, len(req.Recievers))
	for idx, val := range req.Recievers {
		output[idx] = val.DisplayName
	}
	return conversion.JoinStrings(output)
}

func (req *RequestBody) createMailContent(mediaCountry string, idx int) string {
	receiverName := req.Recievers[idx].DisplayName
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
