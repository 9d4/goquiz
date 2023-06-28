package entity

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/94d/goquiz/util"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
)

func Seed() {
	log.Println("Seeding...")

	reader := viper.New()
	reader.SetConfigName("users")
	reader.SetConfigType("yml")
	reader.AddConfigPath(".")

	err := reader.ReadInConfig()
	if err != nil {
		log.Println("Unable to seed users,", err)
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Make sure to create users.yml; See users.example.yml for example")
		}
	}

	SeedUsers(reader.AllSettings())

	reader.SetConfigName("quiz")
	err = reader.ReadInConfig()
	if err != nil {
		log.Println("Unable to seed quiz,", err)
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Make sure to create quiz.yml; See quiz.example.yml for example")
		}
	}
	SeedQuestions(reader.AllSettings())

	log.Println("Seeding...done")
}

func SeedUsers(data map[string]interface{}) {
	if data["users"] == nil {
		return
	}

	ref := reflect.TypeOf(data["users"])
	if ref.Kind() != reflect.Slice {
		return
	}

	users := data["users"].([]interface{})
	for _, user := range users {
		usr, ok := user.(map[string]interface{})
		if !ok {
			continue
		}

		if usr["username"] == nil || usr["password"] == nil {
			continue
		}
		username, ok := usr["username"].(string)
		if !ok {
			continue
		}
		password, ok := usr["password"].(string)
		if !ok {
			continue
		}
		password, err := util.HashPassword(password)
		if err != nil {
			continue
		}

		if usr["fullname"] == nil {
			usr["fullname"] = ""
		}
		fullname, ok := usr["fullname"].(string)
		if !ok {
			fullname = ""
		}

		user := &User{Fullname: fullname, Username: username, Password: password}
		DB().Save(user)
	}
}

func SeedQuestions(data map[string]interface{}) {
	name, ok := data["name"].(string)
	if !ok {
		log.Printf("Invalid quiz name, should be string; got %#v", data["name"])
		return
	}
	SaveQuizName(name)

	ref := reflect.TypeOf(data["questions"])
	if ref.Kind() != reflect.Slice {
		log.Printf("Invalid questions, should be array of question")
		return
	}

	questions := data["questions"].([]interface{})
	for _, q := range questions {
		q, ok := q.(map[string]interface{})
		if !ok {
			continue
		}

		body, ok := q["body"].(string)
		if !ok {
			continue
		}

		question := &Question{
			Body: body,
		}
		tx, err := DB().Begin(true)
		if err != nil {
			log.Println(err)
			continue
		}

		if err := tx.Save(question); err != nil {
			log.Println(err)
			continue
		}

		if reflect.TypeOf(q["choices"]).Kind() != reflect.Slice {
			tx.Rollback()
			continue
		}

		choices, ok := q["choices"].([]interface{})
		if !ok {
			tx.Rollback()
			continue
		}

		for _, c := range choices {
			ch, ok := c.(map[string]interface{})
			if !ok {
				continue
			}

			body, ok := ch["body"].(string)
			if !ok {
				continue
			}
			correct, _ := ch["correct"].(bool)

			choice := &Choice{
				QuestionID: question.ID,
				Body:       body,
				Correct:    correct,
			}
			tx.Save(choice)
		}

		tx.Commit()
	}
}

func SeedQuestionExcel() {
	f, err := excelize.OpenFile("quiz.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	if f.SheetCount < 1 {
		log.Fatal("invalid: sheet less than 1")
	}

	sheetName := f.GetSheetName(0)
	rows, err := f.Rows(sheetName)
	if err != nil {
		log.Fatal(err)
	}

	// save sheetName as the quiz name
	SaveQuizName(sheetName)

	head := &header{}

	rowIndex := 0
	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			fmt.Println(err)
		}

		// Building the header
		if rowIndex == 0 {
			for i, colCell := range row {

				if str := strings.ToLower(colCell); str == "number" {
					head.SetNumberIndex(i)
					continue
				}

				if str := strings.ToLower(colCell); str == "question" {
					head.SetQuestionIndex(i)
					continue
				}

				if str := strings.ToLower(colCell); str == "correct" {
					head.SetCorrectIndex(i)
					continue
				}

				if str := strings.Split(colCell, " "); len(str) > 1 && strings.ToLower(str[0]) == "option" {
					head.AddToOptions(str[1], i)
					continue
				}
			}
			rowIndex++
			continue
		}

		if !head.Complete() {
			log.Fatal("header invalid")
		}

		tx, err := DB().Begin(true)
		if err != nil {
			log.Fatal(err)
		}

		question := &Question{}
		question.Number = strings.TrimSpace(row[*head.NumberIndex])
		question.Body = strings.TrimSpace(row[*head.QuestionIndex])
		tx.Save(question)

		correct := strings.Split(row[*head.CorrectIndex], ",")
		for correctStr, correctIndex := range head.GetOptions() {
			colCell := strings.TrimSpace(row[correctIndex])
			if colCell != "" {
				ch := Choice{
					QuestionID: question.ID,
					Body:       colCell,
				}

				// Check if choice is correct
				for _, correctCol := range correct {
					if correctCol == correctStr {
						ch.Correct = true
						break
					}
				}

				tx.Save(&ch)
			}
		}

		tx.Commit()
		rowIndex++
	}
	if err = rows.Close(); err != nil {
		fmt.Println(err)
	}
}

type header struct {
	NumberIndex   *int
	QuestionIndex *int

	// OptionsIndex contains indices of choices
	// The column looks like: Option 1 | Option 2 | Option A | Correct  etc.
	// So when iterating the row, and it's time for "Correct" column which contains like "1,2,A"
	// then we can split by comma(,) and check each correct options in OptionsIndex["1"] in current row.
	OptionsIndex *map[string]int
	CorrectIndex *int
}

func (h *header) Complete() bool {
	if h.NumberIndex == nil || h.QuestionIndex == nil || h.OptionsIndex == nil || h.CorrectIndex == nil {
		return false
	}

	return true
}

func (h *header) SetNumberIndex(i int) {
	if h.NumberIndex == nil {
		h.NumberIndex = new(int)
	}
	*h.NumberIndex = i
}

func (h *header) SetQuestionIndex(i int) {
	if h.QuestionIndex == nil {
		h.QuestionIndex = new(int)
	}
	*h.QuestionIndex = i
}

func (h *header) SetCorrectIndex(i int) {
	if h.CorrectIndex == nil {
		h.CorrectIndex = new(int)
	}
	*h.CorrectIndex = i
}

func (h *header) AddToOptions(value string, index int) {
	if h.OptionsIndex == nil {
		h.OptionsIndex = new(map[string]int)
		*h.OptionsIndex = make(map[string]int)
	}
	(*h.OptionsIndex)[value] = index
}

func (h *header) IsIndexInOptions(index int) bool {
	if h.OptionsIndex == nil {
		return false
	}
	for _, v := range *h.OptionsIndex {
		if v == index {
			return true
		}
	}
	return false
}

func (h *header) GetOptions() map[string]int {
	return *h.OptionsIndex
}
