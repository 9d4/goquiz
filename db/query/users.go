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
	insertUserStmt *sql.Stmt

	prepareUserStmtsOnce sync.Once
)

func prepareUserStmts(db *sql.DB) (err error) {
	insertUserStmt, err = db.Prepare("INSERT INTO users (fullname, username, password) VALUES (?,?,?)")
	if err != nil {
		return err
	}
	defer insertUserStmt.Close()

	return nil
}

func InsertUser(db *sql.DB, usr *User) (res sql.Result, err error) {
	prepareUserStmtsOnce.Do(func() { err = prepareUserStmts(db) })
	if err != nil {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		return
	}

	res, err = tx.Stmt(insertUserStmt).Exec(usr.Fullname, usr.Username, usr.Password)
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	return
}
