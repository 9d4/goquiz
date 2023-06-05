package db

import (
	"database/sql"
	"log"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	DBName      = "file:goquiz.sqlite3?cache=shared&mode=rwc"
	db          *sql.DB
	connectOnce sync.Once
)

func DB() *sql.DB {
	return db
}

func Connect() (err error) {
	connectOnce.Do(func() {
		err = connect(DBName)
	})
	return
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
