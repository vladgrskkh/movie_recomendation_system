package mailer

import (
	"bytes"
	"context"
	"embed"
	"text/template"
	"time"

	"github.com/mailersend/mailersend-go"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	dialer *mailersend.Mailersend
	sender string
}

func NewMailer(APIKey, sender string) Mailer {
	dialer := mailersend.NewMailersend(APIKey)

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	from := mailersend.From{
		Email: m.sender,
	}

	recipientMailer := []mailersend.Recipient{
		{
			Email: recipient,
		},
	}

	msg := m.dialer.Email.NewMessage()
	msg.SetFrom(from)
	msg.SetRecipients(recipientMailer)
	msg.SetSubject(subject.String())
	msg.SetText(plainBody.String())
	msg.SetHTML(htmlBody.String())

	for i := 1; i <= 3; i++ {
		_, err = m.dialer.Email.Send(ctx, msg)
		if nil == err {
			return nil
		}

		time.Sleep(1500 * time.Millisecond)
	}

	return err
}
