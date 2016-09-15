package main

import (
	"bytes"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/image/webp"
)

var schema = `
CREATE TABLE IF NOT EXISTS users (
	id             serial       PRIMARY KEY,
	username       citext       UNIQUE NOT NULL,
	email          citext       NOT NULL,
	full_name      varchar(100) NOT NULL,
	hash           varchar(255) NOT NULL,
	reset_token    varchar(255) NOT NULL,
	key            varchar(255) NOT NULL,
	created_at     timestamp    WITH TIME ZONE NOT NULL,
	signed_up_at   timestamp    WITH TIME ZONE NOT NULL,
	verified_at    timestamp    WITH TIME ZONE NOT NULL,
	admin          boolean      NOT NULL,
	comment_notify boolean NOT NULL
);
CREATE TABLE IF NOT EXISTS uploads (
	id          varchar(22)   PRIMARY KEY,
	title       varchar(255)  NOT NULL,
	user_id     integer       NOT NULL,
	description varchar(4095) NOT NULL,
	posted_at   timestamp     WITH TIME ZONE NOT NULL,
	approved    boolean       NOT NULL
);
CREATE TABLE IF NOT EXISTS tags (
	upload_id   varchar(22) NOT NULL,
	tag         citext      NOT NULL,
	radius	    real        NOT NULL,
	label_angle real        NOT NULL,
	x           real        NOT NULL,
	y           real        NOT NULL
);
CREATE TABLE IF NOT EXISTS view_counts (
	upload_id  varchar(22) NOT NULL,
	ip_address inet        NOT NULL,
	UNIQUE (upload_id, ip_address)
);
CREATE TABLE IF NOT EXISTS comments (
	id        serial        PRIMARY KEY,
	user_id   integer       NOT NULL,
	upload_id varchar(22)   NOT NULL,
	comment   varchar(4095) NOT NULL,
	posted_at timestamp     WITH TIME ZONE NOT NULL
);
CREATE INDEX IF NOT EXISTS view_counts_upload_id_index ON view_counts (upload_id);
`

type Comment struct {
	Id           int       `db:"id"`
	UserId       int       `db:"user_id"`
	UploadId     string    `db:"upload_id"`
	Comment      string    `db:"comment"`
	PostedAt     time.Time `db:"posted_at"`
	Author       User
	FormatedDate string
}

type User struct {
	Id            int       `db:"id"`
	Username      string    `db:"username"`
	Email         string    `db:"email"`
	FullName      string    `db:"full_name"`
	Hash          string    `db:"hash"`
	ResetToken    string    `db:"reset_token"`
	Key           string    `db:"key"`
	CreatedAt     time.Time `db:"created_at"`
	SignedUpAt    time.Time `db:"signed_up_at"`
	VerifiedAt    time.Time `db:"verified_at"`
	Admin         bool      `db:"admin"`
	CommentNotify bool      `db:"comment_notify"`
}

type Upload struct {
	Id           string    `db:"id"`
	Title        string    `db:"title"`
	UserId       int       `db:"user_id"`
	Description  string    `db:"description"`
	PostedAt     time.Time `db:"posted_at"`
	Approved     bool      `db:"approved"`
	Author       User
	FormatedDate string
}

type Page struct {
	Title   string
	File    string
	Data    interface{}
	Content template.HTML
}

var db *sqlx.DB

var store *sessions.CookieStore

func init() {
	keys, _ := ioutil.ReadFile("cookie_keys.txt")
	split := strings.Split(string(keys), "\n")
	store = sessions.NewCookieStore(
		fromHex(split[0]),
		fromHex(split[1]),
	)
}

func getSession(r *http.Request) *sessions.Session {
	session, _ := store.Get(r, "session")
	return session
}

func ip(r *http.Request) string {
	return r.Header.Get("X-FORWARDED-FOR")
}

func formatDate(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05+00:00")
}

func main() {
	log.Println("Booting up...")
	rand.Seed(time.Now().UnixNano())
	var err error
	log.Printf("Connecting to database... ")
	db, err = sqlx.Connect("postgres", "dbname=clearskies sslmode=disable")
	if err != nil {
		log.Print(err)
	}
	log.Printf("done!\n")
	db.MustExec(schema)
	router := mux.NewRouter()
	router.HandleFunc("/", homePageHandler)
	router.HandleFunc("/account", accountPageHandler).
		Methods("GET")
	router.HandleFunc("/changepassword", changePasswordPageHandler).
		Methods("GET")
	router.HandleFunc("/edit/{Id:[a-zA-Z0-9]+}", editPageHandler).
		Methods("GET")
	router.HandleFunc("/login", loginPageHandler).
		Methods("GET")
	router.HandleFunc("/logout", logoutPageHandler).
		Methods("GET")
	router.HandleFunc("/recover", recoveryPageHandler).
		Methods("GET")
	router.HandleFunc("/reset/{Token}", resetPageHandler).
		Methods("GET")
	router.HandleFunc("/sendverification", sendVerificationHandler).
		Methods("GET")
	router.HandleFunc("/signup", signupPageHandler).
		Methods("GET")
	router.HandleFunc("/tags/{Id:[a-zA-Z0-9]+}", getTagsHandler).
		Methods("GET")
	router.HandleFunc("/thumbnails/{Id:[a-zA-Z0-9]+}", thumbnailHandler).
		Methods("GET")
	router.HandleFunc("/upload", uploadPageHandler).
		Methods("GET")
	router.HandleFunc("/user/{Username}", userPageHandler).
		Methods("GET")
	router.HandleFunc("/verify/{EmailCode}", emailVerifyHandler).
		Methods("GET")
	router.HandleFunc("/view/{Id:[a-zA-Z0-9]+}", viewPageHandler).
		Methods("GET")
	router.HandleFunc("/account/save", saveSettingsHandler).
		Methods("POST")
	router.HandleFunc("/approve/{Id:[a-zA-Z0-9]+}", approvalHandler).
		Methods("POST")
	router.HandleFunc("/changepassword", changePasswordHandler).
		Methods("POST")
	router.HandleFunc("/cleartags/{Id:[a-zA-Z0-9]+}", clearTagsHandler).
		Methods("POST")
	router.HandleFunc("/comment/{Id:[a-zA-Z0-9]+}", commentHandler).
		Methods("POST")
	router.HandleFunc("/delete/{Id:[a-zA-Z0-9]+}", deleteHandler).
		Methods("POST")
	router.HandleFunc("/deletecomment/{Id:[0-9]+}", deleteCommentHandler).
		Methods("POST")
	router.HandleFunc("/edit/{Id:[a-zA-Z0-9]+}", editHandler).
		Methods("POST")
	router.HandleFunc("/generatetags", generateTagsHandler).
		Methods("POST")
	router.HandleFunc("/login", loginHandler).
		Methods("POST")
	router.HandleFunc("/salt", saltHandler).
		Methods("POST")
	router.HandleFunc("/savetags/{Id:[a-zA-Z0-9]+}", saveTagsHandler).
		Methods("POST")
	router.HandleFunc("/sendpasswordreset", sendPasswordResetHandler).
		Methods("POST")
	router.HandleFunc("/signup", signupHandler).
		Methods("POST")
	router.HandleFunc("/upload", uploadHandler).
		Methods("POST")
	router.PathPrefix("/").HandlerFunc(staticHandler)
	serveMux := http.NewServeMux()
	thumb := regexp.MustCompile(`^/thumbnails/[a-zA-Z0-9]{5}$`)
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !thumb.MatchString(r.RequestURI) {
			log.Print(ip(r), ": ", r.Method, " ", r.RequestURI)
		}
		session := getSession(r)
		if _, ok := session.Values["SignedIn"]; !ok {
			session.Values["SignedIn"] = false
			session.Values["Username"] = ""
		}
		if _, ok := session.Values["Verified"]; !ok {
			session.Values["Verified"] = false
		}
		if _, ok := session.Values["Admin"]; !ok {
			session.Values["Admin"] = false
		}
		user := User{}
		db.Get(&user, "SELECT * FROM users WHERE username = $1", session.Values["Username"])
		context.Set(r, "csrf", string(deriveExpiryCode("CSRF", 0, fromHex(user.Key))))
		session.Options.HttpOnly = true
		session.Save(r, w)
		router.ServeHTTP(w, r)
	})
	server := http.Server{
		Addr:    "127.0.0.1:9090",
		Handler: serveMux,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
	}
	server.ListenAndServe()
}

func (u *User) Verified() bool {
	return !u.VerifiedAt.IsZero()
}

func (p *Page) Render(w http.ResponseWriter) {
	content, err := template.ParseFiles("templates/" + p.File)
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	err = content.Execute(&buf, p.Data)
	if err != nil {
		log.Fatal(err)
	}
	p.Content = template.HTML(buf.Bytes())
	base, _ := template.ParseFiles("templates/base.html")
	w.Header().Set("content-Type", "text/html")
	base.Execute(w, p)
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	uploads := []Upload{}
	r.ParseForm()
	query := r.FormValue("search")
	if query == "" {
		db.Select(&uploads, "SELECT * FROM uploads ORDER BY posted_at DESC")
	} else {
		uploadsByTitle := []Upload{}
		db.Select(&uploadsByTitle, "SELECT * FROM uploads WHERE title ~* ('.*' || $1 || '.*')", query)
		tags := []Tag{}
		db.Select(&tags, "SELECT * FROM tags WHERE tag ~* ('.*' || $1 || '.*')", query)
		uploadsByTag := []Upload{}
		for _, tag := range tags {
			upload := Upload{}
			db.Get(&upload, "SELECT * FROM uploads WHERE id = $1", tag.UploadId)
			uploadsByTag = append(uploadsByTag, upload)
		}
		uploads = append(uploadsByTitle, uploadsByTag...)
		for i := range uploads {
			for j := range uploads[i+1:] {
				if uploads[j].Id == uploads[i].Id {
					uploads[j] = Upload{}
				}
			}
		}
	}
	session := getSession(r)
	p := Page{
		Title: "Home",
		File:  "index.html",
		Data: struct {
			Uploads []Upload
			Query   string
			Session map[interface{}]interface{}
		}{
			uploads,
			query,
			session.Values,
		},
	}
	p.Render(w)
}

func viewPageHandler(w http.ResponseWriter, r *http.Request) {
	upload := Upload{}
	err := db.Get(&upload, "SELECT * FROM uploads WHERE id = $1", mux.Vars(r)["Id"])
	if err != nil {
		log.Print(err)
	}
	err = db.Get(&upload.Author, "SELECT username, full_name FROM users WHERE id = $1", upload.UserId)
	if err != nil {
		log.Print(err)
	}
	session := getSession(r)
	count := 0
	if upload.Id != "" {
		db.Exec("INSERT INTO view_counts (upload_id, ip_address) VALUES ($1, $2)", upload.Id, ip(r))
		db.Get(&count, "SELECT count(*) FROM view_counts WHERE upload_id = $1", upload.Id)
	}
	comments := []Comment{}
	db.Select(&comments, "SELECT * FROM comments WHERE upload_id = $1 ORDER BY posted_at ASC", upload.Id)
	for i := range comments {
		comments[i].FormatedDate = formatDate(comments[i].PostedAt)
		db.Get(&comments[i].Author, "SELECT username FROM users WHERE id = $1", comments[i].UserId)
	}
	upload.FormatedDate = formatDate(upload.PostedAt)
	user := User{}
	db.Get(&user, "SELECT key FROM users WHERE username = $1", session.Values["Username"])
	p := Page{
		Title: upload.Title,
		File:  "view.html",
		Data: struct {
			Upload   Upload
			Count    int
			CSRF     string
			Comments []Comment
			Session  map[interface{}]interface{}
		}{
			upload,
			count,
			string(deriveExpiryCode("CSRF", 0, fromHex(user.Key))),
			comments,
			session.Values,
		},
	}
	p.Render(w)
}

func editPageHandler(w http.ResponseWriter, r *http.Request) {
	upload := Upload{}
	db.Get(&upload, "SELECT * FROM uploads WHERE id = $1", mux.Vars(r)["Id"])
	db.Get(&upload.Author, "SELECT username FROM users WHERE id = $1", upload.UserId)
	session := getSession(r)
	p := Page{
		Title: "Edit",
		File:  "edit.html",
		Data: struct {
			Upload  Upload
			Session map[interface{}]interface{}
			CSRF    string
		}{
			upload,
			session.Values,
			context.Get(r, "csrf").(string),
		},
	}
	p.Render(w)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	id := mux.Vars(r)["Id"]
	upload := Upload{}
	db.Get(&upload, "SELECT * FROM uploads WHERE id = $1", id)
	db.Get(&upload.Author, "SELECT username FROM users WHERE id = $1", upload.UserId)
	if upload.Author.Username != session.Values["Username"] {
		errorMessage(w, r, "Prohibited.")
		return
	}
	user := User{}
	db.Get(&user, "SELECT * FROM users WHERE username = $1", session.Values["Username"])
	if !checkExpiryCode(r.PostFormValue("csrf"), "CSRF", user.Key) {
		errorMessage(w, r, "Bad CSRF token.")
		return
	}
	title := r.PostFormValue("title")
	description := r.PostFormValue("description")
	db.Exec("UPDATE uploads SET (title, description) = ($1, $2) WHERE id = $3", title, description, upload.Id)
	http.Redirect(w, r, "/view/"+id, http.StatusFound)
}

func signupPageHandler(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	p := Page{
		Title: "Signup",
		File:  "signup.html",
		Data: struct {
			Session map[interface{}]interface{}
		}{
			session.Values,
		},
	}
	p.Render(w)
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{
		Title: "Login",
		File:  "login.html",
	}
	session := getSession(r)
	var notice string
	if session.Values["Username"] == "" && session.Values["EmailCode"] != nil {
		notice = "Please log in to complete the verification process."
	}
	p.Data = struct {
		Notice  string
		Session map[interface{}]interface{}
	}{
		notice,
		session.Values,
	}
	p.Render(w)
}

func accountPageHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{
		Title: "Account",
		File:  "account.html",
	}
	user := User{}
	session := getSession(r)
	db.Get(&user, "SELECT id, comment_notify FROM users WHERE username = $1", session.Values["Username"])
	var uploads []Upload
	db.Select(&uploads, "SELECT * FROM uploads WHERE user_id = $1 ORDER BY posted_at DESC", user.Id)
	p.Data = struct {
		Uploads       []Upload
		CommentNotify bool
		Session       map[interface{}]interface{}
	}{
		uploads,
		user.CommentNotify,
		session.Values,
	}
	p.Render(w)
}

func userPageHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{
		Title: mux.Vars(r)["Username"],
		File:  "user.html",
	}
	user := User{}
	db.Get(&user, "SELECT id, username, full_name FROM users WHERE username = $1", mux.Vars(r)["Username"])
	var uploads []Upload
	db.Select(&uploads, "SELECT * FROM uploads WHERE user_id = $1 ORDER BY posted_at DESC", user.Id)
	session := getSession(r)
	p.Data = struct {
		Uploads []Upload
		User    User
		Session map[interface{}]interface{}
	}{
		uploads,
		user,
		session.Values,
	}
	p.Render(w)
}

func uploadPageHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{
		Title: "Upload",
		File:  "upload.html",
	}
	session := getSession(r)
	p.Data = struct {
		Session map[interface{}]interface{}
	}{
		session.Values,
	}
	p.Render(w)
}

func logoutPageHandler(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func recoveryPageHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{
		Title: "Recover",
		File:  "recover.html",
	}
	session := getSession(r)
	p.Data = struct {
		Session map[interface{}]interface{}
	}{
		session.Values,
	}
	p.Render(w)
}

func checkExpiryCode(code string, context string, key string) bool {
	for o := time.Duration(0); o >= -time.Hour; o -= time.Hour {
		if subtle.ConstantTimeCompare([]byte(code), deriveExpiryCode(context, o, fromHex(key))) == 1 {
			return true
		}
	}
	return false
}

func resetPageHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{
		Title: "Change Password",
		File:  "changepassword.html",
	}
	resetToken := mux.Vars(r)["Token"]
	user := User{}
	err := db.Get(&user, "SELECT * FROM users WHERE reset_token = $1", resetToken)
	if err == sql.ErrNoRows {
		log.Println("Reset page handler: No user exists with token: " + resetToken)
		errorMessage(w, r, "")
		return
	}
	if !checkExpiryCode(resetToken, "RESET_TOKEN", user.Key) {
		log.Println("Reset page handler: Invalid token")
		errorMessage(w, r, "")
		return
	}
	session := getSession(r)
	p.Data = struct {
		Username   string
		ResetToken string
		Session    map[interface{}]interface{}
	}{
		user.Username,
		resetToken,
		session.Values,
	}
	p.Render(w)
}

func changePasswordPageHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{
		Title: "Change Password",
		File:  "changepassword.html",
	}
	session := getSession(r)
	p.Data = struct {
		Username   string
		ResetToken string
		Session    map[interface{}]interface{}
	}{
		session.Values["Username"].(string),
		"",
		session.Values,
	}
	p.Render(w)
}

func saveSettingsHandler(w http.ResponseWriter, r *http.Request) {
	commentNotify := (r.PostFormValue("commentNotify") == "true")
	session := getSession(r)
	db.Exec("UPDATE users SET comment_notify = $1 WHERE username = $2", commentNotify, session.Values["Username"])
	http.Redirect(w, r, "/account", http.StatusFound)
}

func commentHandler(w http.ResponseWriter, r *http.Request) {
	upload := Upload{}
	err := db.Get(&upload, "SELECT id, user_id, title FROM uploads WHERE id = $1", mux.Vars(r)["Id"])
	if err == sql.ErrNoRows {
		log.Print("Comment handler: Upload does not exist: ", mux.Vars(r)["Id"])
		errorMessage(w, r, "Upload does not exist!")
		return
	}
	db.Get(&upload.Author, "SELECT id, email, comment_notify FROM users WHERE id = $1", upload.UserId)
	session := getSession(r)
	if !session.Values["Verified"].(bool) {
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
	user := User{}
	db.Get(&user, "SELECT id, username, key FROM users WHERE username = $1", session.Values["Username"])
	comment := r.PostFormValue("comment")
	if !checkExpiryCode(r.PostFormValue("csrf"), "CSRF", user.Key) {
		log.Print("COmment handler: Bad CSRF token")
		errorMessage(w, r, "CSRF tokens have expired, please go back and refresh the page.")
		return
	}
	_, err = db.Exec(`INSERT INTO comments (user_id, upload_id, comment, posted_at)
			 VALUES ($1, $2, $3, $4)`,
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

func deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	comment := Comment{}
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
	session := getSession(r)
	user := User{}
	db.Get(&user, "SELECT id, key FROM users WHERE username = $1", session.Values["Username"])
	if !checkExpiryCode(r.PostFormValue("csrf"), "CSRF", user.Key) {
		log.Print("COmment handler: Bad CSRF token")
		errorMessage(w, r, "CSRF tokens have expired, please go back and refresh the page.")
		return
	}
	if !session.Values["Admin"].(bool) && user.Id != comment.UserId {
		errorMessage(w, r, "Prohibited.")
		return
	}
	db.Exec("DELETE FROM comments WHERE id = $1", comment.Id)
	http.Redirect(w, r, "/view/"+comment.UploadId, http.StatusFound)
}

func changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	resetToken := r.PostFormValue("resetToken")
	oldPasscode := r.PostFormValue("oldPasscode")
	newPasscode := r.PostFormValue("newPasscode")
	user := User{}
	if resetToken != "" {
		err := db.Get(&user, "SELECT * FROM users WHERE reset_token = $1", resetToken)
		if err == sql.ErrNoRows {
			log.Println("Change password handler: Invalid reset token: " + resetToken)
			w.WriteHeader(500)
			return
		}
	} else {
		session := getSession(r)
		if session.Values["Username"] == "" {
			log.Println("Change password handler: Not logged in")
			w.WriteHeader(500)
			return
		} else if err := db.Get(&user, "SELECT * FROM users WHERE username = $1", session.Values["Username"]); err != nil {
			log.Print("Change password handler: ", err)
			w.WriteHeader(500)
			return
		}
		if !checkPasscode(user, oldPasscode) {
			log.Print("Change password handler: Incorrect password for password change attempt")
			w.WriteHeader(500)
			w.Write([]byte("incorrect password"))
			return
		}
	}
	err := db.Get(&user, "SELECT * FROM users WHERE username = $1", username)
	if err == sql.ErrNoRows {
		log.Println("Change password handler: User does not exist: " + username)
		w.WriteHeader(500)
		return
	} else if err != nil {
		log.Print("Change password handler: ", err)
	}
	hashDigest := hash(newPasscode, user.Key)
	db.Exec("UPDATE users SET hash = $1 WHERE username = $2", hashDigest, user.Username)
}

var apiURL string = "https://www.google.com/recaptcha/api/siteverify"

func recaptchaTest(gRecaptchaResponse string) bool {
	s, _ := ioutil.ReadFile("recaptcha_secret.txt")
	secret := string(s)
	v := url.Values{}
	v.Add("secret", secret)
	v.Add("response", gRecaptchaResponse)
	resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(v.Encode()))
	if err != nil {
		log.Println("Recaptcha Test: API failure")
		return false
	}
	body, _ := ioutil.ReadAll(resp.Body)
	log.Print(string(body))
	form := struct {
		Success      bool
		Challenge_ts time.Time
		Hostname     string
	}{}
	json.Unmarshal(body, &form)
	return form.Success
}

func errorMessage(w http.ResponseWriter, r *http.Request, errMsg string) {
	p := Page{
		Title: "500",
		File:  "error.html",
	}
	session := getSession(r)
	p.Data = struct {
		ErrorMessage string
		Session      map[interface{}]interface{}
	}{
		errMsg,
		session.Values,
	}
	p.Render(w)
	w.WriteHeader(500)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	if !session.Values["Verified"].(bool) {
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
		tryId := randBase57String(5)
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
	user := User{}
	db.Get(&user, "SELECT id FROM users WHERE username = $1", session.Values["Username"])
	db.Exec(`INSERT INTO uploads (id, title, user_id, description, posted_at, approved)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		id, title, user.Id, description, time.Now().UTC(), false,
	)
	http.Redirect(w, r, "/view/"+id, http.StatusFound)
	sendNewUploadNotification(id)
}

func sendNewUploadNotification(id string) {
	upload := Upload{}
	db.Get(&upload, "SELECT title, user_id FROM uploads WHERE id = $1", id)
	user := User{}
	db.Get(&user, "SELECT username FROM users WHERE id = $1", upload.UserId)
	link, _ := url.Parse("https://clearskies.space")
	link.Path = "/view/" + id
	go mail.Send("clearskies.space@gmail.com", upload.Title+" by "+user.Username, link.String())
}

func approvalHandler(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	if !session.Values["Admin"].(bool) {
		return
	}
	id := mux.Vars(r)["Id"]
	db.Exec("UPDATE uploads SET approved = $1 WHERE id = $2", true, id)
	http.Redirect(w, r, "/", http.StatusFound)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["Id"]
	log.Print(r.RequestURI)
	log.Print(mux.Vars(r)["Id"])
	upload := Upload{}
	err := db.Get(&upload, "SELECT user_id FROM uploads WHERE id = $1", id)
	if err == sql.ErrNoRows {
		errorMessage(w, r, "Image doesn't exist in the first place.")
		return
	} else if err != nil {
		errorMessage(w, r, "")
		return
	}
	author := User{}
	db.Get(&author, "SELECT username FROM users WHERE id = $1", upload.UserId)
	session := getSession(r)
	if !session.Values["Admin"].(bool) && session.Values["Username"] != author.Username {
		errorMessage(w, r, "Prohibited.")
		return
	}
	db.Exec("DELETE FROM uploads WHERE id = $1", id)
	db.Exec("DELETE FROM comments WHERE upload_id = $1", id)
	os.Remove("static/uploads/" + id)
	os.Remove("static/thumbnails/" + id)
	http.Redirect(w, r, "/view/"+id, http.StatusFound)
}

func decodeImage(r io.Reader, format string) (img image.Image) {
	switch format {
	case "gif":
		img, _ = gif.Decode(r)
	case "jpeg":
		img, _ = jpeg.Decode(r)
	case "png":
		img, _ = png.Decode(r)
	case "webp":
		img, _ = webp.Decode(r)
	}
	return
}

func thumbnailHandler(w http.ResponseWriter, r *http.Request) {
	filename := "static/thumbnails/" + mux.Vars(r)["Id"]
	root := http.Dir("static/thumbnails")
	httpFile, err := root.Open(mux.Vars(r)["Id"])
	if err == nil {
		io.Copy(w, httpFile)
		return
	}
	filename = "static/uploads/" + mux.Vars(r)["Id"]
	f, _ := os.Open(filename)
	config, format, _ := image.DecodeConfig(f)
	f.Seek(0, 0)
	img := decodeImage(f, format)
	if img == nil {
		w.WriteHeader(404)
		w.Write([]byte("404 page not found"))
		return
	}
	var dim uint = 300
	var width uint = 0
	var height uint = dim
	if config.Height > config.Width {
		width = dim
		height = 0
	}
	img = resize.Resize(width, height, img, resize.MitchellNetravali)
	img, _ = cutter.Crop(img, cutter.Config{
		Width:  int(dim),
		Height: int(dim),
		Mode:   cutter.Centered,
	})
	var jpegBytes bytes.Buffer
	jpeg.Encode(&jpegBytes, img, &jpeg.Options{
		Quality: 90,
	})
	w.Write(jpegBytes.Bytes())
	f, _ = os.Create("static/thumbnails/" + mux.Vars(r)["Id"])
	f.Write(jpegBytes.Bytes())
}

func hash(passcode, userKey string) string {
	clientHash := fromHex(passcode)
	serverSalt := deriveKey("SERVER_SALT", 16, fromHex(userKey))
	N, R, P := 1<<14, 8, 1
	serverHash, _ := scrypt.Key(clientHash, serverSalt, N, R, P, 16)
	return fmt.Sprintf(
		"|s|%d|%d|%d|%d|%s",
		N, R, P, len(serverSalt), toHex(serverHash),
	)
}

func (u *User) Valid() bool {
	return (u.Username != "" || u.Email != "" && u.Verified()) &&
		u.Hash != "" &&
		!u.SignedUpAt.IsZero() &&
		len(u.Hash) >= 22 &&
		validEmail(u.Email) &&
		validUsername(u.Username) &&
		validHexKey(u.Key)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	passcode := r.PostFormValue("passcode")
	if !validEmail(email) {
		log.Println("Signup handler: Invalid email address: " + email)
		w.WriteHeader(500)
		return
	}
	if !validUsername(username) {
		log.Println("Signup handler: Invalid username.")
		w.WriteHeader(500)
		return
	}
	user := User{}
	err := db.Get(&user, "SELECT * FROM users WHERE username = $1", username)
	if err == sql.ErrNoRows {
		log.Println("Signup handler: Handshake failure")
		w.WriteHeader(500)
		return
	}
	if user.Valid() {
		log.Println("Signup handler: Username already exists: " + username)
		w.WriteHeader(500)
		w.Write([]byte("username exists"))
		return
	}
	users := []User{}
	err = db.Select(&users, "SELECT * FROM users WHERE email = $1", email)
	for i := range users {
		if users[i].Verified() {
			log.Println("Signup handler: Verified email already exists: " + email)
			w.WriteHeader(500)
			w.Write([]byte("email exists"))
			return
		}
	}
	if time.Since(user.CreatedAt).Seconds() > 30 {
		log.Println("Signup handler: Handshake timeout")
		w.WriteHeader(500)
		return
	}
	if !validHexKey(user.Key) {
		log.Println("Signup handler: Handshake failure: Bad user key: " + user.Key)
		w.WriteHeader(500)
		return
	}
	hashDigest := hash(passcode, user.Key)
	db.Exec(`
		UPDATE users
		SET (email, hash, signed_up_at) = ($1, $2, $3)
		WHERE username = $4`,
		email, hashDigest, time.Now(), username,
	)
	go sendEmailVerification(username)
}

func checkPasscode(user User, passcode string) bool {
	if len(user.Hash) < 2 {
		log.Println("Login failure: Missing hash")
		return false
	}
	sep := string(user.Hash[0])
	alg := string(user.Hash[1])
	if alg != "s" {
		log.Fatal("Unknown hash algorithm: " + alg)
	}
	hashArgs := strings.Replace(user.Hash[2:], sep, " ", -1)
	var N, R, P, saltSize int
	var hashString string
	fmt.Sscanf(hashArgs, "%d%d%d%d%s", &N, &R, &P, &saltSize, &hashString)
	serverSalt := deriveKey("SERVER_SALT", saltSize, fromHex(user.Key))
	hash := fromHex(hashString)
	t := time.Now()
	hashToVerify, _ := scrypt.Key(fromHex(passcode), serverSalt, N, R, P, len(hash))
	log.Print(time.Since(t))
	if subtle.ConstantTimeCompare(hash, hashToVerify) == 0 {
		log.Println("Login failure: Hashes do not match")
		return false
	}
	return true
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	usernameOrEmail := r.PostFormValue("usernameOrEmail")
	passcode := r.PostFormValue("passcode")
	if !validUsername(usernameOrEmail) && !validEmail(usernameOrEmail) {
		log.Println("Invalid username or email: " + usernameOrEmail)
		w.WriteHeader(500)
		return
	}
	if !validHexKey(passcode) {
		log.Println("Invalid passcode: " + passcode)
		w.WriteHeader(500)
		return
	}
	user := User{}
	var err error
	var username, email string
	if validEmail(usernameOrEmail) {
		email = usernameOrEmail
		users := []User{}
		err = db.Select(&users, "SELECT * FROM users WHERE email = $1", email)
		if len(users) > 1 {
			var i int
			for i = 0; i < len(users); i++ {
				if users[i].Verified() {
					user = users[i]
					break
				}
			}
			if i == len(users) {
				w.WriteHeader(500)
				w.Write([]byte("multiple unverified emails"))
				return
			}
		} else if len(users) == 1 {
			user = users[0]
		}
		username = user.Username
	} else {
		username = usernameOrEmail
		err = db.Get(&user, "SELECT * FROM users WHERE username = $1", username)
		email = user.Email
	}
	if err == sql.ErrNoRows {
		log.Println("Login failure: Username or email does not exist: ", usernameOrEmail)
		w.WriteHeader(500)
		return
	}
	if !checkPasscode(user, passcode) {
		w.WriteHeader(500)
		return
	}
	session := getSession(r)
	session.Values["SignedIn"] = true
	session.Values["Username"] = username
	if session.Values["EmailCode"] != nil {
		err := verifyEmail(username, session.Values["EmailCode"].(string))
		if err != nil {
			log.Print(err)
		}
		session.Values["Verified"] = (err == nil)
		delete(session.Values, "EmailCode")
	} else {
		session.Values["Verified"] = user.Verified()
	}
	session.Values["Admin"] = ("xunatai" == user.Username)
	session.Save(r, w)
}

func saltHandler(w http.ResponseWriter, r *http.Request) {
	usernameOrEmail := r.PostFormValue("usernameOrEmail")
	if !validUsername(usernameOrEmail) && !validEmail(usernameOrEmail) {
		log.Println("Salt handler: Invalid username or email: " + usernameOrEmail)
		w.WriteHeader(500)
		return
	}
	user := User{}
	var err error
	var username, email string
	if validEmail(usernameOrEmail) {
		email = usernameOrEmail
		users := []User{}
		err = db.Select(&users, "SELECT * FROM users WHERE email = $1", email)
		if len(users) > 1 {
			i := 0
			for i := 0; i < len(users); i++ {
				if users[i].Verified() {
					user = users[i]
					break
				}
			}
			if i == len(users) {
				w.WriteHeader(500)
				w.Write([]byte("multiple unverified emails"))
				return
			}
		} else if len(users) == 1 {
			user = users[0]
		}
		username = user.Username
	} else {
		username = usernameOrEmail
		err = db.Get(&user, "SELECT * FROM users WHERE username = $1", username)
		email = user.Email
	}
	var key []byte
	if err == sql.ErrNoRows {
		key = cryptoRand(16)
		db.Exec(
			`INSERT INTO users (
				username,
				email,
				full_name,
				hash,
				reset_token,
				key,
				created_at,
				signed_up_at,
				verified_at,
				admin
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			username, email, "", "", "", toHex(key), time.Now().UTC(), time.Time{}, time.Time{}, false, true,
		)
	} else {
		key = fromHex(user.Key)
	}
	if !user.Valid() {
		db.Exec("UPDATE users SET created_at = $1 WHERE username = $2", time.Now().UTC(), username)
	}
	w.Header().Set("content-Type", "text/plain")
	w.Write([]byte(toHex(deriveKey("CLIENT_SALT", 16, key))))
}

func sendPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	gRecaptchaResponse := r.PostFormValue("g-recaptcha-response")
	if !recaptchaTest(gRecaptchaResponse) {
		log.Println("Upload handler: Robot alert!")
		errorMessage(w, r, "You need to prove that you are not a robot!")
		return
	}
	usernameOrEmail := r.PostFormValue("usernameOrEmail")
	user := User{}
	err := db.Get(&user, "SELECT * FROM users WHERE username = $1", usernameOrEmail)
	if err == sql.ErrNoRows {
		err = db.Get(&user, "SELECT * FROM users WHERE email = $1", usernameOrEmail)
		if err == sql.ErrNoRows {
			log.Println("Reset password handler: Username or email does not exist: " + usernameOrEmail)
			errorMessage(w, r, "Username or email does not exist.")
			return
		}
	}
	resetToken := deriveExpiryCode("RESET_TOKEN", 0, fromHex(user.Key))
	db.Exec("UPDATE users SET reset_token = $1 WHERE id = $2", resetToken, user.Id)
	link, _ := url.Parse("https://clearskies.space")
	link.Path = "/reset/" + string(resetToken)
	go mail.Send(user.Email, "Password Reset", "Click the following link to reset your password to ClearSkies.space. If you did NOT initiate this request, then please ignore this email.\n"+link.String())
	http.Redirect(w, r, "/", http.StatusFound)
}

func sendVerificationHandler(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	username := session.Values["Username"].(string)
	sendEmailVerification(username)
	http.Redirect(w, r, "/account", http.StatusFound)
}

func emailVerifyHandler(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	if session.Values["Username"] != "" {
		err := verifyEmail(session.Values["Username"].(string), mux.Vars(r)["EmailCode"])
		if err != nil {
			log.Print("Verify email: ", err)
			if !session.Values["Verified"].(bool) {
				errorMessage(w, r, "Failed to verify email! Please log in and and click resend in your account page.")
			} else {
				w.WriteHeader(404)
				w.Write([]byte("404 page not found"))
			}
			return
		} else {
			session.Values["Verified"] = true
			session.Save(r, w)
			http.Redirect(w, r, "/", http.StatusFound)
		}
	} else {
		session.Values["EmailCode"] = mux.Vars(r)["EmailCode"]
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func deriveExpiryCode(context string, offset time.Duration, key []byte) []byte {
	context += time.Now().UTC().Add(offset).Truncate(time.Hour).String()
	return deriveBase57Key(context, 22, key)
}

func verifyEmail(username, emailCode string) error {
	user := User{}
	err := db.Get(&user, "SELECT * FROM users WHERE username = $1", username)
	if err == sql.ErrNoRows {
		return errors.New("user does not exist")
	}
	if !user.Valid() {
		return errors.New("invalid user")
	}
	verified := false
	for o := time.Duration(0); o >= -time.Hour; o -= time.Hour {
		if subtle.ConstantTimeCompare([]byte(emailCode), deriveExpiryCode("EMAIL_CODE", o, fromHex(user.Key))) == 1 {
			verified = true
			break
		}
	}
	if !verified {
		return errors.New("email code verification failed")
	}
	db.Exec("UPDATE users SET verified_at = $1 WHERE username = $2", time.Now().UTC(), user.Username)
	return nil
}

func sendEmailVerification(username string) error {
	user := User{}
	err := db.Get(&user, "SELECT * FROM users WHERE username = $1", username)
	if err == sql.ErrNoRows {
		return errors.New("user does not exist")
	}
	if !user.Valid() {
		return errors.New("invalid user")
	}
	if user.Verified() {
		return errors.New("email already verified")
	}
	link, _ := url.Parse("https://clearskies.space")
	link.Path = "/verify/" + string(deriveExpiryCode("EMAIL_CODE", 0, fromHex(user.Key)))
	go mail.Send(
		user.Email,
		"Please confirm your email address",
		"You signed up for ClearSkies.space. Please click the following link to confirm your email address.\n"+link.String()+"\nIf you did NOT sign up for this website under the username \""+user.Username+"\" please ignore this email.")
	return nil
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	root := http.Dir("static")
	file, err := root.Open(r.URL.Path)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("404 page not found"))
		return
	} else if stat, _ := file.Stat(); stat.IsDir() {
		_, err := root.Open(r.URL.Path + "/index.html")
		if err != nil {
			w.WriteHeader(404)
			w.Write([]byte("404 page not found"))
			return
		} else {
			r.URL.Path += "/index.html"
		}
	}
	http.FileServer(root).ServeHTTP(w, r)
}
