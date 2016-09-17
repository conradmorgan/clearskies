package session

import (
	"clearskies/app/utils"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

func init() {
	keys, _ := ioutil.ReadFile("config/cookie_keys.txt")
	split := strings.Split(string(keys), "\n")
	store = sessions.NewCookieStore(
		utils.FromHex(split[0]),
		utils.FromHex(split[1]),
	)
}

func Get(r *http.Request) *sessions.Session {
	s, _ := store.Get(r, "session")
	if _, ok := s.Values["SignedIn"]; !ok {
		s.Values["SignedIn"] = false
		s.Values["Username"] = ""
	}
	if _, ok := s.Values["Verified"]; !ok {
		s.Values["Verified"] = false
	}
	if _, ok := s.Values["Admin"]; !ok {
		s.Values["Admin"] = false
	}
	s.Options.HttpOnly = true
	return s
}
