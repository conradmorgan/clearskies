package controller

import (
	"clearskies/app/model"
	"clearskies/app/session"
	"clearskies/app/view"
	"net/http"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	uploads := []model.Upload{}
	r.ParseForm()
	query := r.FormValue("search")
	if query == "" {
		db.Select(&uploads, "SELECT * FROM uploads ORDER BY posted_at DESC")
	} else {
		uploadsByTitle := []model.Upload{}
		db.Select(&uploadsByTitle, "SELECT * FROM uploads WHERE title ~* ('.*' || $1 || '.*')", query)
		tags := []Tag{}
		db.Select(&tags, "SELECT * FROM tags WHERE tag ~* ('.*' || $1 || '.*')", query)
		uploadsByTag := []model.Upload{}
		for _, tag := range tags {
			upload := model.Upload{}
			db.Get(&upload, "SELECT * FROM uploads WHERE id = $1", tag.UploadId)
			uploadsByTag = append(uploadsByTag, upload)
		}
		uploads = append(uploadsByTitle, uploadsByTag...)
		for i := range uploads {
			for j := range uploads[i+1:] {
				if uploads[j].Id == uploads[i].Id {
					uploads[j] = model.Upload{}
				}
			}
		}
	}
	s := session.Get(r)
	v := view.New("index.html", "Home")
	v.Vars["Uploads"] = uploads
	v.Vars["Query"] = query
	v.Vars["Session"] = s.Vars
	v.Render(w)
}
