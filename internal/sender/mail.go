package sender

import (
	"encoding/json"
	"net/smtp"

	"github.com/avraam311/delayed-notifier/internal/models/domain"
)

type Mail struct {
	host     string
	port     string
	auth     smtp.Auth
	from     string
	password string
}

func NewMail(host, port, user, from, password string) *Mail {
	auth := smtp.PlainAuth("", user, password, host)
	return &Mail{
		host:     host,
		port:     port,
		auth:     auth,
		from:     from,
		password: password,
	}
}

func (m *Mail) SendMessage(msg []byte) error {
	not := &domain.Notification{}
	json.Unmarshal(msg, not)
	to := []string{not.Mail}
	msgToSend := []byte("Subject: Notifying about your to do\r\n" +
		"\r\n" +
		not.Message + "\r\n")

	err := smtp.SendMail(m.host+":"+m.port, m.auth, m.from, to, msgToSend)
	if err != nil {
		return err
	}
	return nil
}
