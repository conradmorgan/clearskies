package database

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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

var Db *sqlx.DB

func init() {
	log.Printf("Connecting to database... ")
	var err error
	Db, err = sqlx.Connect("postgres", "dbname=clearskies sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("success!\n")
	Db.MustExec(schema)
}
