package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"sync"

	"github.com/gopl-dev/server/app"
)

var initDriverOnce sync.Once
var driver Sender

const (
	SMTPDriver = "smtp"
	TestDriver = "test"
)

//go:embed *.html
var templateFiles embed.FS
var templates = template.Must(template.ParseFS(templateFiles, "*.html"))

type Sender interface {
	Send(to string, c Composer) error
}

type Composer interface {
	Subject() string
	TemplateName() string
	Variables() map[string]any
}

func Send(to string, c Composer) (err error) {
	initDriverOnce.Do(func() {
		conf := app.Config().Email
		switch conf.Driver {
		case SMTPDriver:
			driver, err = NewSMTPSender()
		case TestDriver:
			driver = new(TestSender)
		default:
			err = fmt.Errorf("invalid email driver '%s'", conf.Driver)
		}
	})

	if err != nil {
		return
	}

	return driver.Send(to, c)
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
