package mailtips

import (
	"sync"

	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	sgmail "github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendMail via. sendgrid
func SendMail(client *sendgrid.Client, mailReq *MailReq, ch chan<- *rest.Response, wg *sync.WaitGroup) error {
	go func() error {
		for _, reciever := range mailReq.Receivers {
			wg.Add(1)
			from := sgmail.NewEmail(mailReq.From.DisplayName, mailReq.From.Email)
			subject := mailReq.Subject
			to := sgmail.NewEmail(reciever.Email, reciever.Email)
			content := mailReq.Content
			htmlContent := "<h3> A tip for you!" + mailReq.Content + "</h3>"
			message := sgmail.NewSingleEmail(from, subject, to, content, htmlContent)
			resp, err := client.Send(message)
			if err != nil {
				return err
			}
			ch <- resp
		}
		wg.Done()
		return nil
	}()
	wg.Wait()
	return nil
}
