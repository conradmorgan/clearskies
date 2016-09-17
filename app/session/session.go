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

type Session struct {
	session *sessions.Session
	request *http.Request
}

func Get(r *http.Request) *Session {
	session, _ := store.Get(r, "session")
	s := Session{session, r}
	if _, ok := s.Vars()["SignedIn"]; !ok {
		s.Vars()["SignedIn"] = false
		s.Vars()["Username"] = ""
	}
	if _, ok := s.Vars()["Verified"]; !ok {
		s.Vars()["Verified"] = false
	}
	if _, ok := s.Vars()["Admin"]; !ok {
		s.Vars()["Admin"] = false
	}
	s.session.Options.HttpOnly = true
	return &s
}

func (s *Session) Vars() map[interface{}]interface{} {
	return s.session.Values
}

func (s *Session) Save(w http.ResponseWriter) {
	s.session.Save(s.request, w)
}

func (s *Session) Delete() {
	s.session.Options.MaxAge = -1
}
