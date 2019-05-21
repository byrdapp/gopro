package mailtips

import (
	"fmt"
	"log"
	"sync"

	"github.com/sendgrid/sendgrid-go"
	sgmail "github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendMail via. sendgrid
func SendMail(client *sendgrid.Client, mailReq *MailReq, wg *sync.WaitGroup) {
	for _, reciever := range mailReq.Receivers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			from := sgmail.NewEmail(mailReq.From.DisplayName, mailReq.From.Email)
			subject := mailReq.Subject
			to := sgmail.NewEmail(reciever.DisplayName, reciever.Email)
			content := mailReq.Content
			htmlContent := "<h3> A tip for you!" + mailReq.Content + "</h3>"
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
