package session

import (
	"clearskies/app/config"
	"clearskies/app/utils"
	"clearskies/app/validation"
	"net/http"

	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

func init() {
	var authKey, encKey []byte
	if !validation.ValidHexKey(config.Sessions.AuthKey) {
		authKey = []byte(config.Sessions.AuthKey)
	} else {
		authKey = utils.FromHex(config.Sessions.AuthKey)
	}
	if !validation.ValidHexKey(config.Sessions.EncKey) {
		encKey = []byte(config.Sessions.EncKey)
	} else {
		encKey = utils.FromHex(config.Sessions.EncKey)
	}
	store = sessions.NewCookieStore(authKey, encKey)
}

type Session struct {
	Vars    map[interface{}]interface{}
	session *sessions.Session
	request *http.Request
}

func Get(r *http.Request) *Session {
	session, _ := store.Get(r, "session")
	s := Session{session: session, request: r}
	s.Vars = session.Values
	if _, ok := s.Vars["SignedIn"]; !ok {
		s.Vars["SignedIn"] = false
		s.Vars["Username"] = ""
	}
	if _, ok := s.Vars["Verified"]; !ok {
		s.Vars["Verified"] = false
	}
	if _, ok := s.Vars["Admin"]; !ok {
		s.Vars["Admin"] = false
	}
	s.session.Options.HttpOnly = true
	return &s
}

func (s *Session) Save(w http.ResponseWriter) {
	s.session.Save(s.request, w)
}

func (s *Session) Delete() {
	s.session.Options.MaxAge = -1
}
