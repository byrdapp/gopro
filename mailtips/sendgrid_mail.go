package mailtips

import (
	"encoding/json"
	"io"
	"os"

	"github.com/sendgrid/rest"
	sendgrid "github.com/sendgrid/sendgrid-go"
	sgmail "github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendMail via. sendgrid
func SendMail(body io.Reader) (*rest.Response, error) {
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API"))
	client.Method = "POST"

	// Decode body to struct
	mail := &MailBody{}
	err := json.NewDecoder(body).Decode(mail)
	if err != nil {
		return nil, err
	}
	// Set SG struct values to JSON receiver values in own built struct
	sgmail := &sgmail.SGMailV3{
		Subject: mail.Subject,
		From:    (*sgmail.Email)(mail.From),
		Personalizations: ([]*sgmail.Personalization{{
			To: castEmail(mail.To),
		}}),
		Content: castContent(mail.Content),
	}

	// // Encode to JSON for sendgrid client
	mailJSON, err := json.Marshal(mail)
	if err != nil {
		return nil, err
	}
	client.Body = mailJSON
	resp, err := client.Send(sgmail)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func castContent(input []*Content) []*sgmail.Content {
	output := make([]*sgmail.Content, len(input))
	for index, content := range input {
		output[index] = (*sgmail.Content)(content)
	}
	return output
}

func castEmail(input []*Email) []*sgmail.Email {
	output := make([]*sgmail.Email, len(input))
	for index, content := range input {
		output[index] = (*sgmail.Email)(content)
	}
	return output
}
