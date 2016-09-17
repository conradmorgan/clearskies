package controller

import (
	"clearskies/app/model"
	"clearskies/app/session"
	"clearskies/app/view"
	"net/http"

	"github.com/gorilla/mux"
)

func UserPage(w http.ResponseWriter, r *http.Request) {
	user := model.User{}
	db.Get(&user, "SELECT id, username, full_name FROM users WHERE username = $1", mux.Vars(r)["Username"])
	var uploads []model.Upload
	db.Select(&uploads, "SELECT * FROM uploads WHERE user_id = $1 ORDER BY posted_at DESC", user.Id)
	s := session.Get(r)
	v := view.New("user.html", mux.Vars(r)["Username"])
	v.Vars["Uploads"] = uploads
	v.Vars["User"] = user
	v.Vars["Session"] = s.Vars
	v.Render(w)
}
