package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"strings"
)

type Mailer struct {
	auth    smtp.Auth
	address string
	server  string
}

var mail Mailer

func init() {
	host := "smtp.gmail.com"
	s, _ := ioutil.ReadFile("mail_login.txt")
	split := strings.Split(string(s), "\n")
	username := split[0]
	password := split[1]

	mail = Mailer{
		auth:    smtp.PlainAuth("", username, string(password), host),
		address: username,
		server:  host + ":587",
	}
}

func (m *Mailer) Send(to, subject, msg string) {
	body := makeBody(to, subject, msg)

	log.Print(string(body))
	err := smtp.SendMail(
		m.server, m.auth, m.address, []string{to}, body,
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
