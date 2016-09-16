package controller

import (
	"clearskies/app/session"
	"net/http"

	"github.com/gorilla/mux"
)

func ApproveUpload(w http.ResponseWriter, r *http.Request) {
	session := session.Get(r)
	if !session.Values["Admin"].(bool) {
		return
	}
	id := mux.Vars(r)["Id"]
	db.Exec("UPDATE uploads SET approved = $1 WHERE id = $2", true, id)
	http.Redirect(w, r, "/", http.StatusFound)
}
