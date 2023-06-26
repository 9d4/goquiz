package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/94d/goquiz/entity"
	"github.com/94d/goquiz/util"
	"github.com/94d/goquiz/web"
	"github.com/asdine/storm/q"
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
		return
	}

	if status != "started" {
		w.WriteHeader(http.StatusTooEarly)
		w.Write([]byte("Quiz hasn't started yet"))
		return
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

func (s *server) handleQuizAnswer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Answer string `json:"answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	usr, err := getUser(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var status string
	err = entity.QuizGet(fmt.Sprintf("%s:status", usr.Username), &status)
	if err != nil && !errors.Is(err, entity.ErrNotFound) {
		w.WriteHeader(http.StatusInternalServerError)
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

	var question entity.Question
	if s.db.One("ID", questionIDs[cursor], &question) != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var choice entity.Choice
	if err := s.db.Select(q.And(q.Eq("ID", req.Answer), q.Eq("QuestionID", question.ID))).First(&choice); err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Ilegal answer"))
		return
	}

	entity.QuizSet(usr.Username+":cursor", (1 + cursor))

	choiceID, err := strconv.ParseInt(req.Answer, 10, 0)
	if err != nil {
		log.Println(err)
		return
	}

	var userChoice entity.Choice
	if err := s.db.One("ID", choiceID, &userChoice); err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Ilegal answer"))
		return
	}

	ans := entity.Answer{
		ChoiceID:   userChoice.ID,
		Correct:    userChoice.Correct,
		UserID:     usr.ID,
		QuestionID: question.ID,
	}
	s.db.Save(&ans)

	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleQuizFinish(w http.ResponseWriter, r *http.Request) {
	usr, err := getUser(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var status string
	err = entity.QuizGet(fmt.Sprintf("%s:status", usr.Username), &status)
	if err != nil && !errors.Is(err, entity.ErrNotFound) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if status == "finished" {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("Quiz already finished"))
		return
	}

	if status != "started" {
		w.WriteHeader(http.StatusTooEarly)
		w.Write([]byte("Quiz hasn't started yet"))
		return
	}

	var cursor int
	entity.QuizGet(usr.Username+":cursor", &cursor)

	if cursor < entity.CountQuestions() {
		w.WriteHeader(http.StatusTooEarly)
		w.Write([]byte("Please answer all questions first!"))
		return
	}

	entity.QuizSet(usr.Username+":status", "finished")
}

func (s *server) handleQuizResult(w http.ResponseWriter, r *http.Request) {
	usr, err := getUser(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var userStatus string
	entity.QuizGet(usr.Username+":status", &userStatus)

	if userStatus != "finished" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Please finish the quiz first!"))
		return
	}

	var answers []entity.Answer
	entity.DB().Find("UserID", usr.ID, &answers)

	score, correct := calculateScore(answers, entity.CountQuestions())

	s.JSON(w, map[string]interface{}{
		"score":         score,
		"totalQuestion": entity.CountQuestions(),
		"correctAnswer": correct,
	})
}

func (s *server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	tmpl := template.New("index")
	tmpl, err := tmpl.Funcs(template.FuncMap{
		"add": func(i, j int) int {
			return i + j
		},
	}).Parse(string(web.Dashboard()))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	data := s.resourceAdmin()

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func (s *server) handleAdminRaw(w http.ResponseWriter, r *http.Request) {
	s.JSON(w, s.resourceAdmin())
}

func (s *server) resourceAdmin() interface{} {
	type out struct {
		entity.User
		Answers []entity.Answer `json:"answers"`
		Score   entity.Score    `json:"score"`
		Status  string          `json:"status"`
	}
	var output struct {
		Students  []out             `json:"students"`
		Questions []entity.Question `json:"questions"`
	}

	students := []entity.User{}
	entity.DB().All(&students)

	for _, s := range students {
		d := out{
			User:    s,
			Answers: make([]entity.Answer, entity.CountQuestions()),
		}

		var answers []entity.Answer
		entity.DB().Find("UserID", s.ID, &answers)

		// score calc
		d.Score.UserID = s.ID
		d.Score.Value, _ = calculateScore(answers, entity.CountQuestions())

		for _, a := range answers {
			d.Answers[a.QuestionID-1] = a
		}

		var status string
		entity.QuizGet(s.Username+":status", &status)

		if entity.Onlines.Check(s.ID) {
			d.Status = string(entity.StatusOnline)

			if status == "started" {
				d.Status = string(entity.StatusWorking)
			}
		} else {
			d.Status = string(entity.StatusOffline)
		}

		output.Students = append(output.Students, d)
	}

	questions := []entity.Question{}
	entity.DB().All(&questions)

	output.Questions = questions

	return output
}

func calculateScore(answers []entity.Answer, question int) (score float64, correct int) {
	correctAnswers := []entity.Answer{}
	for _, a := range answers {
		if a.Correct {
			correctAnswers = append(correctAnswers, a)
		}
	}

	return float64(len(correctAnswers)*100) / float64(question), len(correctAnswers)
}
