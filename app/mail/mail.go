package mail

import (
	"clearskies/app/config"
	"fmt"
	"log"
	"net/smtp"
	"strconv"
)

type mailer struct {
	auth    smtp.Auth
	address string
	server  string
}

var mail mailer

func init() {
	mail = mailer{
		auth: smtp.PlainAuth(
			"", config.Mail.Username, config.Mail.Password, config.Mail.Host,
		),
		address: config.Mail.Username,
		server:  config.Mail.Host + ":" + strconv.Itoa(config.Mail.Port),
	}
}

func Send(to, subject, msg string) {
	body := makeBody(to, subject, msg)

	log.Print(string(body))
	err := smtp.SendMail(
		mail.server, mail.auth, mail.address, []string{to}, body,
	)
	if err != nil {
		log.Print(err)
	}
}

func makeBody(to, subject, msg string) (body []byte) {
	format := "To: %s\r\nSubject: %s\r\n\r\n%s"
	body = []byte(fmt.Sprintf(format, to, subject, msg))
	return
}
