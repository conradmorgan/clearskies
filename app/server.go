package app

import (
	"bytes"
	"clearskies/app/config"
	"clearskies/app/database"
	"clearskies/app/model"
	"clearskies/app/session"
	"clearskies/app/utils"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/context"
	"github.com/jmoiron/sqlx"
)

func ip(r *http.Request) string {
	return r.Header.Get("X-FORWARDED-FOR")
}

var db *sqlx.DB

func init() {
	db = database.Db
}

func Serve() {
	go func() {
		db.Exec("DELETE FROM failed_login_attempts")
		time.Sleep(24 * time.Hour)
	}()
	router := routes()
	thumb := regexp.MustCompile(`^/thumbnails/[a-zA-Z0-9]{5}$`)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Filter out thumbnail spam.
		if !thumb.MatchString(r.RequestURI) {
			var body []byte
			// Don't parse and log request bodies that are expected to be large.
			if !(r.RequestURI == "/upload" && r.Method == "POST") {
				// Read r.Body and parse form for logging form values.
				if r.Body != nil {
					body, _ = ioutil.ReadAll(r.Body)
				}
				r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
				r.ParseForm()
			}
			log.Print(ip(r), ": ", r.Method, " ", r.RequestURI, " ", r.Form)
			if !(r.RequestURI == "/upload" && r.Method == "POST") {
				// Restore original state.
				r.Form = nil
				r.PostForm = nil
				r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			}
		}
		s := session.Get(r)
		user := model.User{}
		db.Get(&user, "SELECT * FROM users WHERE username = $1", s.Vars["Username"])
		context.Set(r, "csrf", string(utils.DeriveExpiryCode("CSRF", 0, utils.FromHex(user.Key))))
		s.Save(w)
		router.ServeHTTP(w, r)
	})
	log.Print("Listening on port ", config.Server.Port, "...")
	http.ListenAndServe(
		config.Server.Address+":"+strconv.Itoa(config.Server.Port),
		nil,
	)
}
