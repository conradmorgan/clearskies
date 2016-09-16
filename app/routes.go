package app

import (
	"clearskies/app/controller"

	"github.com/gorilla/mux"
)

func routes() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", controller.HomePage).
		Methods("GET")
	r.HandleFunc("/account", controller.AccountPage).
		Methods("GET")
	r.HandleFunc("/changepassword", controller.ChangePasswordPage).
		Methods("GET")
	r.HandleFunc("/edit/{Id:[a-zA-Z0-9]+}", controller.EditPage).
		Methods("GET")
	r.HandleFunc("/login", controller.LoginPage).
		Methods("GET")
	r.HandleFunc("/logout", controller.Logout).
		Methods("GET")
	r.HandleFunc("/recover", controller.RecoveryPage).
		Methods("GET")
	r.HandleFunc("/reset/{Token}", controller.ResetPage).
		Methods("GET")
	r.HandleFunc("/sendverification", controller.SendVerification).
		Methods("GET")
	r.HandleFunc("/signup", controller.SignupPage).
		Methods("GET")
	r.HandleFunc("/thumbnails/{Id:[a-zA-Z0-9]+}", controller.Thumbnail).
		Methods("GET")
	r.HandleFunc("/upload", controller.UploadPage).
		Methods("GET")
	r.HandleFunc("/user/{Username}", controller.UserPage).
		Methods("GET")
	r.HandleFunc("/verify/{EmailCode}", controller.Verify).
		Methods("GET")
	r.HandleFunc("/view/{Id:[a-zA-Z0-9]+}", controller.ViewPage).
		Methods("GET")
	r.HandleFunc("/account/save", controller.SaveSettings).
		Methods("POST")
	r.HandleFunc("/approve/{Id:[a-zA-Z0-9]+}", controller.ApproveUpload).
		Methods("POST")
	r.HandleFunc("/changepassword", controller.ChangePassword).
		Methods("POST")
	r.HandleFunc("/cleartags/{Id:[a-zA-Z0-9]+}", controller.ClearTags).
		Methods("POST")
	r.HandleFunc("/comment/{Id:[a-zA-Z0-9]+}", controller.Comment).
		Methods("POST")
	r.HandleFunc("/delete/{Id:[a-zA-Z0-9]+}", controller.Delete).
		Methods("POST")
	r.HandleFunc("/deletecomment/{Id:[0-9]+}", controller.DeleteComment).
		Methods("POST")
	r.HandleFunc("/edit/{Id:[a-zA-Z0-9]+}", controller.Edit).
		Methods("POST")
	//r.HandleFunc("/generatetags", controller.GenerateTags).
	//	Methods("POST")
	r.HandleFunc("/login", controller.Login).
		Methods("POST")
	r.HandleFunc("/salt", controller.Salt).
		Methods("POST")
	r.HandleFunc("/savetags/{Id:[a-zA-Z0-9]+}", controller.SaveTags).
		Methods("POST")
	r.HandleFunc("/sendpasswordreset", controller.SendPasswordReset).
		Methods("POST")
	r.HandleFunc("/signup", controller.Signup).
		Methods("POST")
	r.HandleFunc("/upload", controller.Upload).
		Methods("POST")
	r.PathPrefix("/").HandlerFunc(controller.Static)
	return r
}
