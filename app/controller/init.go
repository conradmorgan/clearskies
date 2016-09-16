package controller

import (
	"clearskies/app/database"

	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func init() {
	db = database.Db
}
