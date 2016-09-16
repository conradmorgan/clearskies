package controller

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var apiURL string = "https://www.google.com/recaptcha/api/siteverify"

func recaptchaTest(gRecaptchaResponse string) bool {
	s, _ := ioutil.ReadFile("config/recaptcha_secret.txt")
	secret := string(s)
	v := url.Values{}
	v.Add("secret", secret)
	v.Add("response", gRecaptchaResponse)
	resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(v.Encode()))
	if err != nil {
		log.Println("Recaptcha Test: API failure")
		return false
	}
	body, _ := ioutil.ReadAll(resp.Body)
	log.Print(string(body))
	form := struct {
		Success      bool
		Challenge_ts time.Time
		Hostname     string
	}{}
	json.Unmarshal(body, &form)
	return form.Success
}
