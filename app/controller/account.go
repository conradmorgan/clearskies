package controller

import (
	"clearskies/app/model"
	"clearskies/app/session"
	"clearskies/app/view"
	"net/http"
)

func AccountPage(w http.ResponseWriter, r *http.Request) {
	v := view.View{
		Title: "Account",
		File:  "account.html",
	}
	user := model.User{}
	session := session.Get(r)
	db.Get(&user, "SELECT id, comment_notify FROM users WHERE username = $1", session.Vars()["Username"])
	var uploads []model.Upload
	db.Select(&uploads, "SELECT * FROM uploads WHERE user_id = $1 ORDER BY posted_at DESC", user.Id)
	v.Data = struct {
		Uploads       []model.Upload
		CommentNotify bool
		Session       map[interface{}]interface{}
	}{
		uploads,
		user.CommentNotify,
		session.Vars(),
	}
	v.Render(w)
}

func SaveSettings(w http.ResponseWriter, r *http.Request) {
	commentNotify := (r.PostFormValue("commentNotify") == "true")
	session := session.Get(r)
	db.Exec("UPDATE users SET comment_notify = $1 WHERE username = $2", commentNotify, session.Vars()["Username"])
	http.Redirect(w, r, "/account", http.StatusFound)
}
