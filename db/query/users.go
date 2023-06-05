package query

import (
	"database/sql"
	"sync"
)

type User struct {
	Fullname string
	Username string
	Password string
}

var (
	insertStmt *sql.Stmt

	prepareOnce sync.Once
)

func prepareAllStatements(db *sql.DB) (err error) {
	insertStmt, err = db.Prepare("INSERT INTO users (fullname, username, password) VALUES (?,?,?)")
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	return nil
}

func InsertUser(db *sql.DB, usr *User) (res sql.Result, err error) {
	prepareOnce.Do(func() { err = prepareAllStatements(db) })
	if err != nil {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		return
	}

	res, err = tx.Stmt(insertStmt).Exec(usr.Fullname, usr.Username, usr.Password)
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	return
}
