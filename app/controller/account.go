package controller

import (
	"clearskies/app/model"
	"clearskies/app/session"
	"clearskies/app/view"
	"net/http"
)

func AccountPage(w http.ResponseWriter, r *http.Request) {
	user := model.User{}
	s := session.Get(r)
	db.Get(&user, "SELECT id, comment_notify FROM users WHERE username = $1", s.Vars["Username"])
	var uploads []model.Upload
	db.Select(&uploads, "SELECT * FROM uploads WHERE user_id = $1 ORDER BY posted_at DESC", user.Id)
	v := view.New("account.html", "Account")
	v.Vars["Uploads"] = uploads
	v.Vars["CommentNotify"] = user.CommentNotify
	v.Vars["Session"] = s.Vars
	v.Render(w)
}

func SaveSettings(w http.ResponseWriter, r *http.Request) {
	commentNotify := (r.PostFormValue("commentNotify") == "true")
	s := session.Get(r)
	db.Exec("UPDATE users SET comment_notify = $1 WHERE username = $2", commentNotify, s.Vars["Username"])
	http.Redirect(w, r, "/account", http.StatusFound)
}
