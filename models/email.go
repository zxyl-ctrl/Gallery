package models

import (
	"fmt"

	"github.com/go-mail/mail/v2"
)

const (
	// DefaultSender is the default email address to send emails from.
	DefaultSender = "apiGallery@demomailtrap.com"
	// 使用官方免费提供的域名，因为没钱买域名
)

type EmailService struct {
	// DefaultSender is used as the default sender when one isn't provided for an
	// email. This is also used in functions where the email is a predetermined,
	// like the forgotten password email.
	DefaultSender string

	// unexported fields
	dialer *mail.Dialer
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type Email struct {
	From      string
	To        string
	Subject   string
	Plaintext string
	HTML      string
}

func NewEmailService(config SMTPConfig) *EmailService {
	es := EmailService{
		dialer: mail.NewDialer(
			config.Host,
			config.Port,
			config.Username,
			config.Password,
		),
	}
	return &es
}

func (es *EmailService) Send(email Email) error {
	msg := mail.NewMessage()
	msg.SetHeader("To", email.To)
	// TODO: Set the From field to a default value if it is not
	// set in the Email
	// msg.SetHeader("From", email.From)
	es.setFrom(msg, email)
	// fmt.Println(msg)
	msg.SetHeader("Subject", email.Subject)
	switch {
	case email.Plaintext != "" && email.HTML != "":
		msg.SetBody("text/plain", email.Plaintext)
		msg.AddAlternative("text/html", email.HTML)
	case email.Plaintext != "":
		msg.SetBody("text/plain", email.Plaintext)
	case email.HTML != "":
		msg.SetBody("text/html", email.HTML)
	}
	// fmt.Println(msg)
	err := es.dialer.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

// Usedtosetthesenderofthemessage.Thepriorityis:
// -email.From
// -EmailService.DefaultSender
// -DefaultSender(package const)
func (es *EmailService) setFrom(msg *mail.Message, email Email) {
	var from string
	switch {
	case email.From != "":
		from = email.From
	case es.DefaultSender != "":
		from = es.DefaultSender
	default:
		from = DefaultSender
	}
	msg.SetHeader("From", from)
}

func (es *EmailService) ForgotPassword(to, resetURL string) error {
	email := Email{
		Subject:   "Reset your password",
		To:        to,
		Plaintext: "To reset your password, please visit the following link: " + resetURL,
		HTML:      `<p>To reset your password, please visit the following link <a href="` + resetURL + `</a></p>`,
	}
	fmt.Println(email) // 这里有一个问题，发送邮件可能由于promail.com的筛查导致链接无法显示
	err := es.Send(email)
	if err != nil {
		return fmt.Errorf("forgot password email: %w", err)
	}
	return nil
}
