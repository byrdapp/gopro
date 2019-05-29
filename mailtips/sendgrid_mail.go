package mailtips

import (
	"fmt"
	"log"
	"sync"

	"github.com/sendgrid/sendgrid-go"
	sgmail "github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendMail via. sendgrid
func SendMail(client *sendgrid.Client, body *MailReq, wg *sync.WaitGroup) {
	for _, reciever := range body.Receivers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			from := sgmail.NewEmail(body.From.DisplayName, body.From.Email)
			subject := body.Subject
			to := sgmail.NewEmail(reciever.DisplayName, reciever.Email)
			content := body.Content
			htmlContent := "<h3> A tip for you!" + body.Content + "</h3>"
			message := sgmail.NewSingleEmail(from, subject, to, content, htmlContent)
			resp, err := client.Send(message)
			if err != nil {
				log.Panicf("Error sending mail: %s", err)
			}
			fmt.Printf("%s", resp.Body)
		}()
		wg.Wait()
	}
}
