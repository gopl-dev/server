package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"sync"

	"github.com/gopl-dev/server/app"
	"github.com/wneessen/go-mail"
)

const (
	SMTPDriver = "smtp"
	TestDriver = "test"
)

//go:embed *.html
var templateFiles embed.FS
var templates = template.Must(template.ParseFS(templateFiles, "*.html"))

var (
	smtpSender = new(SMTPSender)
	testSender = new(TestSender)
)

type Composer interface {
	Subject() string
	TemplateName() string
	Variables() map[string]any
}

func Send(to string, c Composer) (err error) {
	switch app.Config().Email.Driver {
	case SMTPDriver:
		err = smtpSender.Send(to, c)
	case TestDriver:
		err = testSender.Send(to, c)
	default:
		err = fmt.Errorf("invalid email driver '%s'", app.Config().Email.Driver)
	}

	return err
}

type Sender interface {
	Send(to string, c Composer) error
}

type SMTPSender struct{}

func (SMTPSender) Send(to string, c Composer) (err error) {
	body, err := renderTemplate(c)
	if err != nil {
		return err
	}

	conf := app.Config().Email
	message := mail.NewMsg()
	err = message.From(conf.From)
	if err != nil {
		return err
	}
	err = message.To(to)
	if err != nil {
		return err
	}

	message.Subject("gopl: " + c.Subject())
	message.SetBodyString(mail.TypeTextHTML, body)

	client, err := mail.NewClient(conf.Host,
		mail.WithPort(conf.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(conf.Username),
		mail.WithPassword(conf.Password),
		mail.WithDebugLog(),
	)
	if err != nil {
		log.Printf("failed to create new mail delivery client: %s", err)
		return err
	}
	err = client.DialAndSend(message)
	if err != nil {
		log.Printf("failed to deliver mail: %s", err)
		return err
	}

	return nil
}

type TestSender struct {
	emails sync.Map
}

func (t *TestSender) Send(to string, c Composer) (err error) {
	body, err := renderTemplate(c)
	if err != nil {
		return err
	}

	t.emails.Store(to, body)
	return nil
}

type TemplateData struct {
	Subject string
	Body    template.HTML
}

func renderTemplate(c Composer) (result string, err error) {
	var buff bytes.Buffer
	err = templates.ExecuteTemplate(&buff, c.TemplateName()+".html", c.Variables())
	if err != nil {
		return
	}

	var tBuff bytes.Buffer
	err = templates.ExecuteTemplate(&tBuff, "template.html", TemplateData{
		Subject: c.Subject(),
		Body:    template.HTML(buff.String()),
	})
	if err != nil {
		return
	}

	return tBuff.String(), nil
}
