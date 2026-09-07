package services

import (
	"fmt"
	"log"
	"mime"
	"net/smtp"
	"strings"

	"awebo/app/infrastructure/config"
)

// Mailer abstracts "sending the one-time code somewhere" so local/dev setups
// don't need real SMTP credentials. Set MAIL_DRIVER=smtp to send for real.
type Mailer interface {
	SendVerificationCode(email, code string) error
}

func NewMailer(cfg *config.Config) Mailer {
	// На тестовом стенде почту не отправляем даже если настроен реальный
	// SMTP — код просто уходит в лог, чтобы можно было пройти вход без
	// доступа к почтовому ящику.
	if cfg.TestServer {
		return consoleMailer{}
	}
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

	ttlMinutes := int(m.cfg.CodeTTL.Minutes())
	if ttlMinutes < 1 {
		ttlMinutes = 1
	}

	msg := buildVerificationEmail(m.cfg.SMTPFrom, email, code, ttlMinutes)

	var auth smtp.Auth
	if m.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPassword, m.cfg.SMTPHost)
	}

	return smtp.SendMail(addr, auth, m.cfg.SMTPFrom, []string{email}, msg)
}

// buildVerificationEmail собирает multipart/alternative письмо (plain-text +
// HTML) с RFC 2047-кодированной темой — без этого кириллица в Subject может
// отображаться битой в некоторых почтовых клиентах.
func buildVerificationEmail(from, to, code string, ttlMinutes int) []byte {
	const boundary = "awebo-boundary-7f3c1a"
	subject := mime.QEncoding.Encode("UTF-8", "Ваш код для входа в awebo")

	var b strings.Builder
	fmt.Fprintf(&b, "From: awebo <%s>\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	b.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	b.WriteString("\r\n")

	fmt.Fprintf(&b, "--%s\r\n", boundary)
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
	fmt.Fprintf(&b,
		"Ваш код для входа: %s\r\nОн действует %d мин.\r\n\r\nЕсли вы не запрашивали вход — просто проигнорируйте это письмо.\r\n\r\n",
		code, ttlMinutes,
	)

	fmt.Fprintf(&b, "--%s\r\n", boundary)
	b.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
	b.WriteString(verificationEmailHTML(code, ttlMinutes))
	b.WriteString("\r\n\r\n")

	fmt.Fprintf(&b, "--%s--\r\n", boundary)

	return []byte(b.String())
}

func verificationEmailHTML(code string, ttlMinutes int) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="ru">
  <body style="margin:0;padding:32px 16px;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0">
      <tr>
        <td align="center">
          <table role="presentation" width="480" cellpadding="0" cellspacing="0" style="max-width:480px;width:100%%;background:#ffffff;border-radius:16px;border:1px solid #e4e4e7;">
            <tr>
              <td style="padding:36px 32px 0 32px;text-align:center;">
                <span style="font-size:22px;font-weight:700;letter-spacing:-0.02em;color:#111827;">awebo</span>
              </td>
            </tr>
            <tr>
              <td style="padding:20px 32px 0 32px;text-align:center;">
                <p style="margin:0;font-size:15px;line-height:1.5;color:#52525b;">
                  Код для входа на awebo
                </p>
              </td>
            </tr>
            <tr>
              <td style="padding:20px 32px;text-align:center;">
                <div style="display:inline-block;padding:16px 28px;background:#f4f4f5;border-radius:12px;font-size:32px;font-weight:700;letter-spacing:0.35em;color:#111827;font-family:'SFMono-Regular',Consolas,Menlo,monospace;">
                  %s
                </div>
              </td>
            </tr>
            <tr>
              <td style="padding:0 32px;text-align:center;">
                <p style="margin:0;font-size:13px;color:#a1a1aa;">
                  Код действует %d мин.
                </p>
              </td>
            </tr>
            <tr>
              <td style="padding:28px 32px 36px 32px;">
                <p style="margin:0;font-size:13px;line-height:1.6;color:#a1a1aa;text-align:center;">
                  Если вы не запрашивали вход — просто проигнорируйте это письмо,<br />никаких действий не потребуется.
                </p>
              </td>
            </tr>
            <tr>
              <td style="padding:20px 32px;background:#fafafa;border-top:1px solid #e4e4e7;border-radius:0 0 16px 16px;text-align:center;">
                <p style="margin:0;font-size:12px;color:#a1a1aa;">awebo — новости, которые важны</p>
              </td>
            </tr>
          </table>
        </td>
      </tr>
    </table>
  </body>
</html>`, code, ttlMinutes)
}
