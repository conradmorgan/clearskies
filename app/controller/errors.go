package controller

import (
	"clearskies/app/session"
	"clearskies/app/view"
	"net/http"
)

func errorMessage(w http.ResponseWriter, r *http.Request, errMsg string) {
	s := session.Get(r)
	v := view.New("error.html", "500")
	v.Vars["ErrorMessage"] = errMsg
	v.Vars["Session"] = s.Vars
	w.WriteHeader(500)
	v.Render(w)
}
