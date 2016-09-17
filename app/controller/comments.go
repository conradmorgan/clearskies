package controller

import (
	"clearskies/app/mail"
	"clearskies/app/model"
	"clearskies/app/session"
	"clearskies/app/utils"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func Comment(w http.ResponseWriter, r *http.Request) {
	upload := model.Upload{}
	err := db.Get(&upload, "SELECT id, user_id, title FROM uploads WHERE id = $1", mux.Vars(r)["Id"])
	if err == sql.ErrNoRows {
		log.Print("Comment handler: Upload does not exist: ", mux.Vars(r)["Id"])
		errorMessage(w, r, "Upload does not exist!")
		return
	}
	db.Get(&upload.Author, "SELECT id, email, comment_notify FROM users WHERE id = $1", upload.UserId)
	s := session.Get(r)
	if !s.Vars()["Verified"].(bool) {
		log.Println("Comment handler: Not verified")
		w.WriteHeader(500)
		return
	}
	gRecaptchaResponse := r.PostFormValue("g-recaptcha-response")
	if !recaptchaTest(gRecaptchaResponse) {
		log.Println("Comment handler: Robot alert!")
		errorMessage(w, r, "You need to prove that you are not a robot!")
		return
	}
	user := model.User{}
	db.Get(&user, "SELECT id, username, key FROM users WHERE username = $1", s.Vars()["Username"])
	comment := r.PostFormValue("comment")
	if !utils.CheckExpiryCode(r.PostFormValue("csrf"), "CSRF", user.Key) {
		log.Print("Comment handler: Bad CSRF token")
		errorMessage(w, r, "CSRF tokens have expired, please go back and refresh the page.")
		return
	}
	_, err = db.Exec(
		"INSERT INTO comments (user_id, upload_id, comment, posted_at) VALUES ($1, $2, $3, $4)",
		user.Id, upload.Id, comment, time.Now().UTC(),
	)
	if err != nil {
		log.Print("Comment handler: ", err)
	}
	if upload.Author.CommentNotify && user.Id != upload.Author.Id {
		go mail.Send(
			upload.Author.Email,
			user.Username+" has commented on your upload",
			"Go to https://clearskies.space/view/"+upload.Id+" to see it.\nIf you do not wish to get these notifications go to your accounts settings at https://clearskies.space/account and disable them.",
		)
	}
	go mail.Send(
		"clearskies.space@gmail.com",
		user.Username+" has commented on "+upload.Title,
		"https://clearskies.space/view/"+upload.Id,
	)
	http.Redirect(w, r, "/view/"+upload.Id, http.StatusFound)
}

func DeleteComment(w http.ResponseWriter, r *http.Request) {
	comment := model.Comment{}
	id, err := strconv.Atoi(mux.Vars(r)["Id"])
	if err != nil {
		log.Print("Delete comment Handler: Invalid comment id: ", mux.Vars(r)["Id"])
		errorMessage(w, r, "Invalid comment id.")
		return
	}
	err = db.Get(&comment, "SELECT * FROM comments WHERE id = $1", id)
	if err == sql.ErrNoRows {
		log.Print("Delete comment Handler: Comment id does not exist: ", mux.Vars(r)["Id"])
		errorMessage(w, r, "Comment does not exist!")
		return
	}
	s := session.Get(r)
	user := model.User{}
	db.Get(&user, "SELECT id, key FROM users WHERE username = $1", s.Vars()["Username"])
	if !utils.CheckExpiryCode(r.PostFormValue("csrf"), "CSRF", user.Key) {
		log.Print("COmment handler: Bad CSRF token")
		errorMessage(w, r, "CSRF tokens have expired, please go back and refresh the page.")
		return
	}
	if !s.Vars()["Admin"].(bool) && user.Id != comment.UserId {
		errorMessage(w, r, "Prohibited.")
		return
	}
	db.Exec("DELETE FROM comments WHERE id = $1", comment.Id)
	http.Redirect(w, r, "/view/"+comment.UploadId, http.StatusFound)
}
