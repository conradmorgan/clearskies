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
	session, _ := store.Get(r, "session")
	return session
}
