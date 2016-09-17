package controller

import (
	"clearskies/app/mail"
	"clearskies/app/model"
	"clearskies/app/session"
	"clearskies/app/utils"
	"clearskies/app/view"
	"database/sql"
	"encoding/json"
	"html/template"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

func ip(r *http.Request) string {
	return r.Header.Get("X-FORWARDED-FOR")
}

func ViewPage(w http.ResponseWriter, r *http.Request) {
	upload := model.Upload{}
	err := db.Get(&upload, "SELECT * FROM uploads WHERE id = $1", mux.Vars(r)["Id"])
	if err != nil {
		log.Print(err)
	}
	err = db.Get(&upload.Author, "SELECT username, full_name FROM users WHERE id = $1", upload.UserId)
	if err != nil {
		log.Print(err)
	}
	count := 0
	if upload.Id != "" {
		db.Exec("INSERT INTO view_counts (upload_id, ip_address) VALUES ($1, $2)", upload.Id, ip(r))
		db.Get(&count, "SELECT count(*) FROM view_counts WHERE upload_id = $1", upload.Id)
	}
	comments := []model.Comment{}
	db.Select(&comments, "SELECT * FROM comments WHERE upload_id = $1 ORDER BY posted_at ASC", upload.Id)
	for i := range comments {
		comments[i].FormatedDate = utils.FormatDate(comments[i].PostedAt)
		db.Get(&comments[i].Author, "SELECT username FROM users WHERE id = $1", comments[i].UserId)
	}
	upload.FormatedDate = utils.FormatDate(upload.PostedAt)
	s := session.Get(r)
	user := model.User{}
	db.Get(&user, "SELECT key FROM users WHERE username = $1", s.Vars["Username"])
	tags := []Tag{}
	db.Select(&tags, "SELECT * FROM tags WHERE upload_id = $1", mux.Vars(r)["Id"])
	tagData, _ := json.Marshal(tags)
    v := view.New("view.html", upload.Title)
    v.Vars["Upload"] = upload
    v.Vars["Count"] = count
    v.Vars["CSRF"] = string(utils.DeriveExpiryCode("CSRF", 0, utils.FromHex(user.Key)))
    v.Vars["Comments"] = comments
    v.Vars["Tags"] = template.JS(tagData)
    v.Vars["Session"] = s.Vars
	v.AddHeader(`
		<meta property="og:url" content="https://clearskies.space/view/{{.Id}}">
		<meta property="og:type" content="website">
		<meta property="og:title" content="{{.Title}} - ClearSkies.space">
		<meta property="og:description" content="{{.Title}}">
		<meta property="og:image" content="https://clearskies.space/uploads/{{.Id}}">`,
		map[string]interface{}{
			"Id":    mux.Vars(r)["Id"],
			"Title": upload.Title,
		},
	)
	v.Render(w)
}

func UploadPage(w http.ResponseWriter, r *http.Request) {
	s := session.Get(r)
	v := view.New("upload.html", "Upload")
	v.Vars["Session"] = s.Vars
	v.Render(w)
}

func Upload(w http.ResponseWriter, r *http.Request) {
	s := session.Get(r)
	if !s.Vars["Verified"].(bool) {
		log.Println("Upload handler: Not verified")
		w.WriteHeader(500)
		return
	}
	r.ParseMultipartForm(25000000)
	title := r.FormValue("title")
	description := r.FormValue("description")
	if len(title) > 255 || len(description) > 4095 {
		log.Println("Upload handler: title or description too long")
		errorMessage(w, r, "Title or description too long!")
		return
	}
	gRecaptchaResponse := r.FormValue("g-recaptcha-response")
	if !recaptchaTest(gRecaptchaResponse) {
		log.Println("Upload handler: Robot alert!")
		errorMessage(w, r, "You need to prove that you are not a robot! Hit back and try again.")
		return
	}
	id := ""
	for i := 0; i < 128; i++ {
		tryId := utils.RandBase57String(5)
		count := 0
		db.Get(&count, "SELECT count(*) FROM uploads WHERE id = $1", id)
		if count == 0 {
			id = tryId
			break
		}
	}
	if id == "" {
		errorMessage(w, r, "Failed to generate an id.")
		return
	}
	f, _, err := r.FormFile("file")
	if err != nil {
		log.Println("Upload handler: No file")
		errorMessage(w, r, "No file chosen.")
		return
	}
	_, format, _ := image.DecodeConfig(f)
	log.Println("Upload request: " + format)
	supported := false
	switch format {
	case "gif":
		supported = true
	case "jpeg":
		supported = true
	case "png":
		supported = true
	case "webp":
		supported = true
	}
	if !supported {
		log.Println("Upload handler: File not supported")
		errorMessage(w, r, "File format not supported. You may upload gif, jpeg, png, or webp.")
		return
	}
	f.Seek(0, 0)
	imageData, _ := ioutil.ReadAll(f)
	ioutil.WriteFile("static/uploads/"+id, imageData, 0644)
	user := model.User{}
	db.Get(&user, "SELECT id FROM users WHERE username = $1", s.Vars["Username"])
	db.Exec(`INSERT INTO uploads (id, title, user_id, description, posted_at, approved)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		id, title, user.Id, description, time.Now().UTC(), false,
	)
	http.Redirect(w, r, "/view/"+id, http.StatusFound)
	sendNewUploadNotification(id)
}

func sendNewUploadNotification(id string) {
	upload := model.Upload{}
	db.Get(&upload, "SELECT title, user_id FROM uploads WHERE id = $1", id)
	user := model.User{}
	db.Get(&user, "SELECT username FROM users WHERE id = $1", upload.UserId)
	link, _ := url.Parse("https://clearskies.space")
	link.Path = "/view/" + id
	go mail.Send("clearskies.space@gmail.com", upload.Title+" by "+user.Username, link.String())
}

func EditPage(w http.ResponseWriter, r *http.Request) {
	upload := model.Upload{}
	db.Get(&upload, "SELECT * FROM uploads WHERE id = $1", mux.Vars(r)["Id"])
	db.Get(&upload.Author, "SELECT username FROM users WHERE id = $1", upload.UserId)
	s := session.Get(r)
	v := view.New("edit.html", "Edit")
	v.Vars["Upload"] = upload
	v.Vars["Session"] = s.Vars
	v.Vars["CSRF"] = context.Get(r, "csrf")
	v.Render(w)
}

func Edit(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["Id"]
	upload := model.Upload{}
	db.Get(&upload, "SELECT * FROM uploads WHERE id = $1", id)
	db.Get(&upload.Author, "SELECT username FROM users WHERE id = $1", upload.UserId)
	s := session.Get(r)
	if upload.Author.Username != s.Vars["Username"] {
		errorMessage(w, r, "Prohibited.")
		return
	}
	user := model.User{}
	db.Get(&user, "SELECT * FROM users WHERE username = $1", s.Vars["Username"])
	if !utils.CheckExpiryCode(r.PostFormValue("csrf"), "CSRF", user.Key) {
		errorMessage(w, r, "Bad CSRF token.")
		return
	}
	title := r.PostFormValue("title")
	description := r.PostFormValue("description")
	db.Exec("UPDATE uploads SET (title, description) = ($1, $2) WHERE id = $3", title, description, upload.Id)
	http.Redirect(w, r, "/view/"+id, http.StatusFound)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["Id"]
	log.Print(r.RequestURI)
	log.Print(mux.Vars(r)["Id"])
	upload := model.Upload{}
	err := db.Get(&upload, "SELECT user_id FROM uploads WHERE id = $1", id)
	if err == sql.ErrNoRows {
		errorMessage(w, r, "Image doesn't exist in the first place.")
		return
	} else if err != nil {
		errorMessage(w, r, "")
		return
	}
	author := model.User{}
	db.Get(&author, "SELECT username FROM users WHERE id = $1", upload.UserId)
	s := session.Get(r)
	if !s.Vars["Admin"].(bool) && s.Vars["Username"] != author.Username {
		errorMessage(w, r, "Prohibited.")
		return
	}
	db.Exec("DELETE FROM uploads WHERE id = $1", id)
	db.Exec("DELETE FROM comments WHERE upload_id = $1", id)
	db.Exec("DELETE FROM view_counts WHERE upload_id = $1", id)
	os.Remove("static/uploads/" + id)
	os.Remove("static/thumbnails/" + id)
	http.Redirect(w, r, "/view/"+id, http.StatusFound)
}
