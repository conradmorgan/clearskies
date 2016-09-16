package controller

import (
	"clearskies/app/session"
	"clearskies/app/view"
	"net/http"
)

func errorMessage(w http.ResponseWriter, r *http.Request, errMsg string) {
	v := view.View{
		Title: "500",
		File:  "error.html",
	}
	s := session.Get(r)
	v.Data = struct {
		ErrorMessage string
		Session      map[interface{}]interface{}
	}{
		errMsg,
		s.Values,
	}
	v.Render(w)
	w.WriteHeader(500)
}
