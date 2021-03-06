package controller

import (
	"clearskies/app/model"
	"clearskies/app/session"
	"clearskies/app/utils"
	"clearskies/app/validation"
	"clearskies/app/view"
	"crypto/subtle"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/scrypt"
)

func LoginPage(w http.ResponseWriter, r *http.Request) {
	s := session.Get(r)
	var notice string
	if s.Vars["Username"] == "" && s.Vars["EmailCode"] != nil {
		notice = "Please log in to complete the verification process."
	}
	var attemptsSlice []int
	db.Select(
		&attemptsSlice,
		`SELECT attempts
		FROM failed_login_attempts
		WHERE ip_address = $1
		ORDER BY attempts DESC`,
		ip(r),
	)
	attempts := 0
	if len(attemptsSlice) >= 1 {
		attempts = attemptsSlice[0]
	}
	v := view.New("login.html", "Log in")
	v.Vars["Notice"] = notice
	v.Vars["Session"] = s.Vars
	v.Vars["IncludeCaptcha"] = (attempts >= 3)
	v.Render(w)
}

func SignupPage(w http.ResponseWriter, r *http.Request) {
	s := session.Get(r)
	v := view.New("signup.html", "Sign up")
	v.Vars["Session"] = s.Vars
	v.Render(w)
}

func Login(w http.ResponseWriter, r *http.Request) {
	usernameOrEmail := r.PostFormValue("usernameOrEmail")
	passcode := r.PostFormValue("passcode")
	if !validation.ValidUsername(usernameOrEmail) && !validation.ValidEmail(usernameOrEmail) {
		log.Println("Invalid username or email: " + usernameOrEmail)
		errorMessage(w, r, "Invalid username or email.")
		return
	}
	user := model.User{}
	var err error
	var username, email string
	if validation.ValidEmail(usernameOrEmail) {
		email = usernameOrEmail
		users := []model.User{}
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
	attempts := 0
	db.Get(&attempts,
		`SELECT attempts
		FROM failed_login_attempts
		WHERE user_id = $1 AND ip_address = $2`,
		user.Id, ip(r),
	)
	session := session.Get(r)
	session.Vars["AttemptedUserId"] = user.Id
	session.Save(w)
	if attempts >= 3 {
		if !recaptchaTest(r.PostFormValue("g-recaptcha-response")) {
			log.Println("Login handler: Robot alert!")
			w.WriteHeader(500)
			w.Write([]byte("recaptcha failure"))
			return
		}
	}
	if !checkPasscode(user, passcode) {
		attempts := 0
		err := db.Get(&attempts,
			`SELECT attempts
			FROM failed_login_attempts
			WHERE user_id = $1 AND ip_address = $2`,
			user.Id, ip(r),
		)
		attempts++
		if err == sql.ErrNoRows {
			_, err := db.Exec(
				`INSERT INTO failed_login_attempts (user_id, ip_address, attempts)
				VALUES ($1, $2, $3)`,
				user.Id, ip(r), 1,
			)
			if err != nil {
				log.Print("Login handler: ", err)
			}
		} else {
			db.Exec(
				`UPDATE failed_login_attempts
				SET attempts = $1
				WHERE user_id = $2 AND ip_address = $3`,
				attempts, user.Id, ip(r),
			)
		}
		if attempts >= 3 {
			w.WriteHeader(500)
			w.Write([]byte("too many failed attempts"))
			return
		}
		w.WriteHeader(500)
		return
	}
	db.Exec(`
		DELETE FROM failed_login_attempts
		WHERE user_id = $1 AND ip_address = $2`,
		user.Id, ip(r),
	)
	session.Vars["SignedIn"] = true
	session.Vars["Username"] = username
	if session.Vars["EmailCode"] != nil {
		session.Vars["Verified"] = checkEmailCode(username, session.Vars["EmailCode"].(string))
		delete(session.Vars, "EmailCode")
	} else {
		session.Vars["Verified"] = user.Verified()
	}
	session.Vars["Admin"] = ("xunatai" == user.Username)
	session.Save(w)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	s := session.Get(r)
	s.Delete()
	s.Save(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	passcode := r.PostFormValue("passcode")
	if !validation.ValidEmail(email) {
		log.Println("Signup handler: Invalid email address: " + email)
		w.WriteHeader(500)
		return
	}
	if !validation.ValidUsername(username) {
		log.Println("Signup handler: Invalid username.")
		w.WriteHeader(500)
		return
	}
	user := model.User{}
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
	users := []model.User{}
	err = db.Select(&users, "SELECT * FROM users WHERE email = $1", email)
	for i := range users {
		if users[i].Verified() {
			log.Println("Signup handler: Verified email already exists: " + email)
			w.WriteHeader(500)
			w.Write([]byte("email exists"))
			return
		}
	}
	if time.Since(user.CreatedAt).Seconds() > 10 {
		log.Println("Signup handler: Handshake timeout")
		w.WriteHeader(500)
		return
	}
	if !validation.ValidHexKey(user.Key) {
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

func Salt(w http.ResponseWriter, r *http.Request) {
	usernameOrEmail := r.PostFormValue("usernameOrEmail")
	if !validation.ValidUsername(usernameOrEmail) && !validation.ValidEmail(usernameOrEmail) {
		log.Println("Salt handler: Invalid username or email: " + usernameOrEmail)
		w.WriteHeader(500)
		return
	}
	user := model.User{}
	var err error
	var username, email string
	if validation.ValidEmail(usernameOrEmail) {
		email = usernameOrEmail
		users := []model.User{}
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
		key = utils.CryptoRand(16)
		_, err := db.Exec(
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
				admin,
				comment_notify
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			username, email, "", "", "", utils.ToHex(key), time.Now().UTC(), time.Time{}, time.Time{}, false, true,
		)
		if err != nil {
			log.Print("Salt handler: ", err)
		}
	} else {
		key = utils.FromHex(user.Key)
	}
	if !user.Valid() {
		db.Exec("UPDATE users SET created_at = $1 WHERE username = $2", time.Now().UTC(), username)
	}
	w.Header().Set("content-Type", "text/plain")
	w.Write([]byte(utils.ToHex(utils.DeriveKey("CLIENT_SALT", 16, key))))

}

func ChangePasswordPage(w http.ResponseWriter, r *http.Request) {
	s := session.Get(r)
	v := view.New("changepassword.html", "Change Password")
	v.Vars["ResetToken"] = ""
	v.Vars["Session"] = s.Vars
	v.Render(w)
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	resetToken := r.PostFormValue("resetToken")
	oldPasscode := r.PostFormValue("oldPasscode")
	newPasscode := r.PostFormValue("newPasscode")
	user := model.User{}
	if resetToken != "" {
		err := db.Get(&user, "SELECT * FROM users WHERE reset_token = $1", resetToken)
		if err == sql.ErrNoRows {
			log.Println("Change password handler: Invalid reset token: " + resetToken)
			w.WriteHeader(500)
			return
		}
	} else {
		s := session.Get(r)
		if s.Vars["Username"] == "" {
			log.Println("Change password handler: Not logged in")
			w.WriteHeader(500)
			return
		} else if err := db.Get(&user, "SELECT * FROM users WHERE username = $1", s.Vars["Username"]); err != nil {
			log.Print("Change password handler: ", err)
			w.WriteHeader(500)
			return
		}
		if !checkPasscode(user, oldPasscode) {
			log.Print("Change password handler: Incorrect password for password change attempt")
			errorMessage(w, r, "Incorrect password.")
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

func hash(passcode, userKey string) string {
	clientHash := utils.FromHex(passcode)
	serverSalt := utils.DeriveKey("SERVER_SALT", 16, utils.FromHex(userKey))
	N, R, P := 1<<14, 8, 1
	serverHash, _ := scrypt.Key(clientHash, serverSalt, N, R, P, 16)
	return fmt.Sprintf(
		"|s|%d|%d|%d|%d|%s",
		N, R, P, len(serverSalt), utils.ToHex(serverHash),
	)
}

func checkPasscode(user model.User, passcode string) bool {
	if len(user.Hash) < 2 {
		log.Print("Login failure: Missing hash for user: ", user.Id)
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
	serverSalt := utils.DeriveKey("SERVER_SALT", saltSize, utils.FromHex(user.Key))
	hash := utils.FromHex(hashString)
	t := time.Now()
	hashToVerify, _ := scrypt.Key(utils.FromHex(passcode), serverSalt, N, R, P, len(hash))
	log.Print(time.Since(t))
	if subtle.ConstantTimeCompare(hash, hashToVerify) == 0 {
		log.Println("Login failure: Hashes do not match")
		return false
	}
	return true
}
