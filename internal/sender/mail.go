package sender

import (
	"encoding/json"
	"net/smtp"

	"github.com/avraam311/delayed-notifier/internal/models"
)

type Mail struct {
	host     string
	port     string
	auth     smtp.Auth
	from     string
	password string
}

func NewMail(host, port, from, password string) *Mail {
	auth := smtp.PlainAuth("", from, password, host)
	return &Mail{
		host:     host,
		port:     port,
		auth:     auth,
		from:     from,
		password: password,
	}
}

func (m *Mail) SendMessage(msg []byte) error {
	not := &models.Notification{}
	json.Unmarshal(msg, not)
	to := not.ToMail
	msgToSend := []byte("Subject: Hello from Go\r\n" +
		"\r\n" +
		not.Message + "\r\n")

	err := smtp.SendMail(m.host+":"+m.port, m.auth, m.from, to, msgToSend)
	if err != nil {
		return err
	}
	return nil
}
