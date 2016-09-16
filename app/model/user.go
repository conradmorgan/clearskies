package model

import (
	"clearskies/app/validation"
	"time"
)

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

func (u *User) Verified() bool {
	return !u.VerifiedAt.IsZero()
}

func (u *User) Valid() bool {
	return (u.Username != "" || u.Email != "" && u.Verified()) &&
		u.Hash != "" &&
		!u.SignedUpAt.IsZero() &&
		len(u.Hash) >= 22 &&
		validation.ValidEmail(u.Email) &&
		validation.ValidUsername(u.Username) &&
		validation.ValidHexKey(u.Key)
}
