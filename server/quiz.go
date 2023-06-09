package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/94d/goquiz/entity"
	"github.com/94d/goquiz/util"
)

func (s *server) handleQuiz(w http.ResponseWriter, r *http.Request) {
	usr, err := getUser(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var status string
	err = entity.QuizGet(fmt.Sprintf("%s:status", usr.Username), &status)
	if err != nil && !errors.Is(err, entity.ErrNotFound) {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if status == "finished" {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("Quiz already finished"))
	}

	if status != "started" {
		w.WriteHeader(http.StatusTooEarly)
		w.Write([]byte("Quiz hasn't started yet"))
	}

	var cursor int
	err = entity.QuizGet(usr.Username+":cursor", &cursor)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var questionIDs []int
	err = entity.QuizGet(usr.Username+":questions", &questionIDs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	questionCount := len(questionIDs)

	if cursor == questionCount {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("All questions answered"))
		return
	}

	var currentQuestion entity.Question
	if s.db.One("ID", questionIDs[cursor], &currentQuestion) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var choices []entity.Choice
	if s.db.Find("QuestionID", currentQuestion.ID, &choices) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var choicesOut []map[string]interface{}
	for _, c := range choices {
		choicesOut = append(choicesOut, map[string]interface{}{
			"id":   c.ID,
			"body": c.Body,
		})
	}

	s.JSON(w, map[string]interface{}{
		"question": map[string]interface{}{
			"id":      currentQuestion.ID,
			"body":    currentQuestion.Body,
			"choices": choicesOut,
		},
	})
}

func (s *server) handleQuizData(w http.ResponseWriter, r *http.Request) {
	s.JSON(w, map[string]interface{}{
		"name": entity.GetQuizName(),
	})
}

func (s *server) handleQuizStart(w http.ResponseWriter, r *http.Request) {
	usr, err := getUser(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var status string
	err = entity.QuizGet(usr.Username+":status", &status)
	if err != nil && !errors.Is(err, entity.ErrNotFound) {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	switch status {
	case "started":
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Quiz already started"))
		return
	case "finished":
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("Quiz already finished"))
		return
	}

	var questions []entity.Question
	if s.db.All(&questions) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	questions = util.Shuffle(questions)

	var questionIDs []int
	for _, q := range questions {
		questionIDs = append(questionIDs, q.ID)
	}

	if entity.QuizSet(usr.Username+":status", "started") != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if entity.QuizSet(usr.Username+":cursor", 0) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if entity.QuizSet(usr.Username+":questions", questionIDs) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.JSON(w, map[string]string{
		"message": "Quiz started",
	})
}
