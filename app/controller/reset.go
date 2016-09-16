package controller

import (
	"clearskies/app/mail"
	"clearskies/app/model"
	"clearskies/app/session"
	"clearskies/app/utils"
	"clearskies/app/view"
	"database/sql"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

func RecoveryPage(w http.ResponseWriter, r *http.Request) {
	v := view.View{
		Title: "Recover",
		File:  "recover.html",
	}
	s := session.Get(r)
	v.Data = struct {
		Session map[interface{}]interface{}
	}{
		s.Values,
	}
	v.Render(w)
}

func ResetPage(w http.ResponseWriter, r *http.Request) {
	v := view.View{
		Title: "Change Password",
		File:  "changepassword.html",
	}
	resetToken := mux.Vars(r)["Token"]
	user := model.User{}
	err := db.Get(&user, "SELECT * FROM users WHERE reset_token = $1", resetToken)
	if err == sql.ErrNoRows {
		log.Println("Reset page handler: No user exists with token: " + resetToken)
		errorMessage(w, r, "")
		return
	}
	if !utils.CheckExpiryCode(resetToken, "RESET_TOKEN", user.Key) {
		errorMessage(w, r, "CSRF tokens have expired. Please go back and refresh the page.")
		return
	}
	s := session.Get(r)
	v.Data = struct {
		Username   string
		ResetToken string
		Session    map[interface{}]interface{}
	}{
		user.Username,
		resetToken,
		s.Values,
	}
	v.Render(w)
}

func SendPasswordReset(w http.ResponseWriter, r *http.Request) {
	gRecaptchaResponse := r.PostFormValue("g-recaptcha-response")
	if !recaptchaTest(gRecaptchaResponse) {
		log.Println("Upload handler: Robot alert!")
		errorMessage(w, r, "You need to prove that you are not a robot!")
		return
	}
	usernameOrEmail := r.PostFormValue("usernameOrEmail")
	user := model.User{}
	err := db.Get(&user, "SELECT * FROM users WHERE username = $1", usernameOrEmail)
	if err == sql.ErrNoRows {
		err = db.Get(&user, "SELECT * FROM users WHERE email = $1", usernameOrEmail)
		if err == sql.ErrNoRows {
			log.Println("Reset password handler: Username or email does not exist: " + usernameOrEmail)
			errorMessage(w, r, "Username or email does not exist.")
			return
		}
	}
	resetToken := utils.DeriveExpiryCode("RESET_TOKEN", 0, utils.FromHex(user.Key))
	db.Exec("UPDATE users SET reset_token = $1 WHERE id = $2", resetToken, user.Id)
	link, _ := url.Parse("https://clearskies.space")
	link.Path = "/reset/" + string(resetToken)
	go mail.Send(user.Email, "Password Reset", "Click the following link to reset your password to ClearSkies.space. If you did NOT initiate this request, then please ignore this email.\n"+link.String())
	http.Redirect(w, r, "/", http.StatusFound)
}
