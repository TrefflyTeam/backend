package mail

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
	"time"
)

type SMTPConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	DefaultFrom  string
}

type Mailer struct {
	dialer      *gomail.Dialer
	defaultFrom string
}

func New(config SMTPConfig) *Mailer {
	dialer := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)
	dialer.TLSConfig = &tls.Config{ServerName: config.Host}
	return &Mailer{
		dialer:      dialer,
		defaultFrom: config.DefaultFrom,
	}
}

func (m *Mailer) SendPasswordReset(to, resetCode string, expiry time.Duration) error {
	subject := "Запрос на сброс пароля"
	expiryMinutes := int(expiry.Minutes())
	plainText := fmt.Sprintf("Сбросить пароль: %s\nДействие истечёт через %d минут", resetCode, expiryMinutes)
	html := fmt.Sprintf(`
        <h1>Сброс пароля</h1>
        <p>Твой код: %s</p>
        <p>Действие истечёт через %d минут</p>
    `, resetCode, expiryMinutes)

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.defaultFrom)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", plainText)
	msg.AddAlternative("text/html", html)

	return m.dialer.DialAndSend(msg)
}