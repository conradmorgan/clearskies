package controller

import (
	"clearskies/app/mail"
	"clearskies/app/model"
	"clearskies/app/session"
	"clearskies/app/utils"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func SendVerification(w http.ResponseWriter, r *http.Request) {
	s := session.Get(r)
	username := s.Vars["Username"].(string)
	sendEmailVerification(username)
	http.Redirect(w, r, "/account", http.StatusFound)
}

func Verify(w http.ResponseWriter, r *http.Request) {
	s := session.Get(r)
	if s.Vars["Username"] != "" {
		ok := checkEmailCode(s.Vars["Username"].(string), mux.Vars(r)["EmailCode"])
		if !ok {
			if !s.Vars["Verified"].(bool) {
				errorMessage(w, r, "Failed to verify email! Please log in and and click resend in your account page.")
			} else {
				w.WriteHeader(404)
				w.Write([]byte("404 page not found"))
			}
			return
		} else {
			s.Vars["Verified"] = true
			s.Save(w)
			http.Redirect(w, r, "/", http.StatusFound)
		}
	} else {
		s.Vars["EmailCode"] = mux.Vars(r)["EmailCode"]
		s.Save(w)
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func sendEmailVerification(username string) error {
	user := model.User{}
	err := db.Get(&user, "SELECT * FROM users WHERE username = $1", username)
	if err == sql.ErrNoRows {
		return errors.New("user does not exist")
	}
	if !user.Valid() {
		return errors.New("invalid user")
	}
	if user.Verified() {
		return errors.New("email already verified")
	}
	emailCode := string(utils.DeriveExpiryCode("EMAIL_CODE", 0, utils.FromHex(user.Key)))
	go mail.Send(
		user.Email,
		"Please confirm your email address",
		fmt.Sprintf(
			`You signed up for ClearSkies.space. Please click the following link to confirm your email address.
			%s
			If you did NOT sign up for this website under the username "%s" please ignore this email.`,
			"https://clearskies.space/verify/"+emailCode,
			user.Username,
		),
	)
	return nil
}

func checkEmailCode(username, emailCode string) bool {
	user := model.User{}
	err := db.Get(&user, "SELECT * FROM users WHERE username = $1", username)
	if err == sql.ErrNoRows {
		return false
	}
	if !user.Valid() {
		return false
	}
	verified := utils.CheckExpiryCode(emailCode, "EMAIL_CODE", user.Key)
	if !verified {
		return false
	}
	db.Exec(
		"UPDATE users SET verified_at = $1 WHERE username = $2",
		time.Now().UTC(), user.Username,
	)
	return true
}
