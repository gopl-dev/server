package email

import (
	"fmt"
	"log"

	"github.com/gopl-dev/server/app"
	"github.com/wneessen/go-mail"
)

// SMTPSender is an implementation Sender interface,
// responsible for sending emails via an external SMTP server.
type SMTPSender struct {
	client *mail.Client
}

// NewSMTPSender creates and initializes a new SMTPSender instance.
func NewSMTPSender() (*SMTPSender, error) {
	conf := app.Config().Email

	client, err := mail.NewClient(conf.Host,
		mail.WithPort(conf.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(conf.Username),
		mail.WithPassword(conf.Password),
		mail.WithDebugLog(),
	)
	if err != nil {
		return nil, err
	}

	return &SMTPSender{
		client: client,
	}, nil
}

// Send composes and dispatches an email message to the specified recipient using the configured SMTP server.
func (s *SMTPSender) Send(to string, c Composer) (err error) {
	body, err := renderTemplate(c)
	if err != nil {
		return err
	}

	conf := app.Config().Email
	message := mail.NewMsg()

	err = message.From(conf.From)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	err = message.To(to)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	message.Subject("gopl: " + c.Subject())
	message.SetBodyString(mail.TypeTextHTML, body)

	err = s.client.DialAndSend(message)
	if err != nil {
		log.Printf("failed to deliver mail: %s", err)

		return err
	}

	return nil
}
