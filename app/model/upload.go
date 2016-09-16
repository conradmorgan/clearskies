package model

import "time"

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
