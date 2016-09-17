package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type server struct {
	Address string
	Port    int
}
type mail struct {
	Host     string
	Port     int
	Username string
	Password string
}
type sessions struct {
	AuthKey string
	EncKey  string
}
type recaptcha struct {
	Secret string
}

var Server server
var Mail mail
var Sessions sessions
var Recaptcha recaptcha

func init() {
	config, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		log.Fatal("Missing config file: ", err)
	}
	err = json.Unmarshal(
		config,
		&struct {
			Server    *server
			Mail      *mail
			Sessions  *sessions
			Recaptcha *recaptcha
		}{
			&Server,
			&Mail,
			&Sessions,
			&Recaptcha,
		},
	)
	log.Printf("\n%#v\n%#v\n%#v\n%#v\n", Server, Mail, Sessions, Recaptcha)
	if err != nil {
		log.Fatal("Bad config file: ", err)
	}
}
