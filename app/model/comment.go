package model

import "time"

type Comment struct {
	Id           int       `db:"id"`
	UserId       int       `db:"user_id"`
	UploadId     string    `db:"upload_id"`
	Comment      string    `db:"comment"`
	PostedAt     time.Time `db:"posted_at"`
	Author       User
	FormatedDate string
}
