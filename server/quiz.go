package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/94d/goquiz/entity"
	"github.com/94d/goquiz/util"
	"github.com/94d/goquiz/web"
	"github.com/asdine/storm"
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

	var answers []string
	if err := entity.QuizGet(usr.Username+":answers", &answers); err != nil {
		if !errors.Is(err, storm.ErrNotFound) {
			return
		}
	}

	answers = append(answers, req.Answer)
	entity.QuizSet(usr.Username+":answers", answers)

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

	// shuffled questions for user
	// contains id of questions
	var userQuestionIDs []int
	if err := entity.QuizGet(usr.Username+":questions", &userQuestionIDs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	questionCount, err := s.db.Count(&entity.Question{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(userQuestionIDs) != questionCount {
		w.WriteHeader(http.StatusTooEarly)
		w.Write([]byte("Please answer all questions first!"))
		log.Println(len(userQuestionIDs))
		log.Println(questionCount)
		return
	}

	var userAnswerIDs []string
	if err := entity.QuizGet(usr.Username+":answers", &userAnswerIDs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	correct := 0
	for i := 0; i < len(userQuestionIDs); i++ {
		var q entity.Question
		s.db.One("ID", userQuestionIDs[i], &q)

		var choices []entity.Choice
		s.db.Find("QuestionID", q.ID, &choices)

		var selChoice entity.Choice
		uai, err := strconv.ParseInt(userAnswerIDs[i], 10, 0)
		if err != nil {
			log.Println(err)
			continue
		}
		s.db.One("ID", uai, &selChoice)

		if selChoice.Correct {
			correct++
		}

		answer := entity.Answer{
			UserID:     usr.ID,
			QuestionID: q.ID,
			ChoiceID:   selChoice.ID,
			Correct:    selChoice.Correct,
		}
		s.db.Save(&answer)
	}

	score := entity.Score{
		UserID: usr.ID,
		Value:  float64(correct*100) / float64(questionCount),
	}
	s.db.Save(&score)

	entity.QuizSet(usr.Username+":status", "finished")
}

func (s *server) handleQuizResult(w http.ResponseWriter, r *http.Request) {
	usr, err := getUser(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var score entity.Score
	if err := s.db.One("UserID", usr.ID, &score); err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Please finish the quiz first!"))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	questionCount, _ := s.db.Count(&entity.Question{})

	var correctAnswers []entity.Answer
	if err := s.db.Select(q.And(q.Eq("UserID", usr.ID), q.Eq("Correct", true))).Find(&correctAnswers); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	s.JSON(w, map[string]interface{}{
		"score":         score.Value,
		"totalQuestion": questionCount,
		"correctAnswer": len(correctAnswers),
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
			Answers: []entity.Answer{},
		}

		entity.DB().Find("UserID", s.ID, &d.Answers)
		entity.DB().One("UserID", s.ID, &d.Score)

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

		sort.SliceStable(d.Answers, func(i, j int) bool {
			return d.Answers[i].QuestionID < d.Answers[j].QuestionID
		})

		output.Students = append(output.Students, d)
	}

	questions := []entity.Question{}
	entity.DB().All(&questions)

	output.Questions = questions

	return output
}
