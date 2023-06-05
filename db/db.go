package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var (
	DBName = "file:goquiz.sqlite3?cache=shared&mode=rwc"
	db     *sql.DB
)

func DB() *sql.DB {
	return db
}

func Connect() error {
	return connect(DBName)
}

func connect(name string) error {
	conn, err := sql.Open("sqlite3", name)
	throwErr(err)

	db = conn

	return nil
}

func throwErr(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
