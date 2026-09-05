package services

import (
	"fmt"
	"log"
	"net/smtp"

	"awebo/app/infrastructure/config"
)

// Mailer abstracts "sending the one-time code somewhere" so local/dev setups
// don't need real SMTP credentials. Set MAIL_DRIVER=smtp to send for real.
type Mailer interface {
	SendVerificationCode(email, code string) error
}

func NewMailer(cfg *config.Config) Mailer {
	if cfg.MailDriver == "smtp" {
		return smtpMailer{cfg: cfg}
	}
	return consoleMailer{}
}

type consoleMailer struct{}

func (consoleMailer) SendVerificationCode(email, code string) error {
	log.Printf("[mail:console] verification code for %s: %s", email, code)
	return nil
}

type smtpMailer struct {
	cfg *config.Config
}

func (m smtpMailer) SendVerificationCode(email, code string) error {
	addr := fmt.Sprintf("%s:%s", m.cfg.SMTPHost, m.cfg.SMTPPort)
	msg := []byte(fmt.Sprintf(
		"To: %s\r\nSubject: Ваш код для входа\r\n\r\nВаш код: %s\r\nОн действует %d секунд.\r\n",
		email, code, int(m.cfg.CodeTTL.Seconds()),
	))

	var auth smtp.Auth
	if m.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPassword, m.cfg.SMTPHost)
	}

	return smtp.SendMail(addr, auth, m.cfg.SMTPFrom, []string{email}, msg)
}
