package query

import (
	"database/sql"
	"sync"
)

var (
	insertQuestionStmt *sql.Stmt
	deleteQuestionStmt *sql.Stmt
	insertChoiceStmt   *sql.Stmt

	prepareQuizStmtsOnce sync.Once
)

type Question struct {
	Body string
}

func prepareQuizStmts(db *sql.DB) (err error) {
	insertQuestionStmt, err = db.Prepare("INSERT INTO questions (body) VALUES (?)")
	if err != nil {
		return err
	}
	defer insertQuestionStmt.Close()

	deleteQuestionStmt, err = db.Prepare("DELETE FROM questions WHERE id=?")
	if err != nil {
		return err
	}
	defer deleteQuestionStmt.Close()

	insertChoiceStmt, err = db.Prepare("INSERT INTO choices (body,correct,question_id) VALUES (?,?,?)")
	if err != nil {
		return err
	}
	defer insertChoiceStmt.Close()

	return nil
}

func InsertQuestion(db *sql.DB, question *Question) (res sql.Result, err error) {
	prepareQuizStmtsOnce.Do(func() { err = prepareQuizStmts(db) })
	if err != nil {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		return
	}

	res, err = tx.Stmt(insertQuestionStmt).Exec(question.Body)
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	return
}

func DeleteQuestionByID(db *sql.DB, id int) (res sql.Result, err error) {
	prepareQuizStmtsOnce.Do(func() { err = prepareQuizStmts(db) })
	if err != nil {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		return
	}

	res, err = tx.Stmt(insertQuestionStmt).Exec(id)
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	return
}

type Choice struct {
	Body       string
	Correct    bool
	QuestionID int
}

func NewChoice(body string, questionID int, correct ...bool) *Choice {
	c := false
	if len(correct) > 0 {
		c = correct[0]
	}

	return &Choice{Body: body, QuestionID: questionID, Correct: c}
}

func InsertChoice(db *sql.DB, choice *Choice) (res sql.Result, err error) {
	prepareQuizStmtsOnce.Do(func() { err = prepareQuizStmts(db) })
	if err != nil {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		return
	}

	res, err = tx.Stmt(insertChoiceStmt).Exec(choice.Body, choice.Correct, choice.QuestionID)
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	return
}
