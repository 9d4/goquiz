package entity

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/94d/goquiz/util"
	"github.com/xuri/excelize/v2"
)

func Seed() {
	start := time.Now()
	log.Println("Seeding...")
	SeedByExcel()
	log.Printf("Seeding...done in %dms\n", time.Since(start).Milliseconds())
}

func SeedByExcel() {
	f, err := excelize.OpenFile("quiz.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	if f.SheetCount < 2 {
		log.Fatal("invalid: sheet less than 2")
	}

	// The sheet of questions should be in the first order
	quizName := f.GetSheetName(0)
	SaveQuizName(quizName)

	questionRows, err := f.Rows(quizName)
	if err != nil {
		log.Fatal(err)
	}
	start := time.Now()
	SeedQuestionExcel(questionRows)
	log.Printf("questions done in %dms\n", time.Since(start).Milliseconds())

	userRows, err := f.Rows("students")
	if err != nil {
		log.Fatal(err)
	}
	start = time.Now()
	SeedUserExcel(userRows)
	log.Printf("students done in %dms\n", time.Since(start).Milliseconds())
}

func SeedUserExcel(rows *excelize.Rows) {
	head := &userHeader{}

	wg := sync.WaitGroup{}
	tx, err := DB().Begin(true)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		wg.Wait()
		tx.Commit()
	}()

	rowIndex := 0
	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			fmt.Println(err)
		}

		// Building the header
		if rowIndex == 0 {
			for i, colCell := range row {
				str := strings.ToLower(strings.TrimSpace(colCell))

				switch str {
				case "fullname":
					head.SetFullnameIndex(i)
				case "username":
					head.SetUsernameIndex(i)
				case "password":
					head.SetPasswordIndex(i)
				}
			}

			rowIndex++
			continue
		}

		if !head.Complete() {
			log.Fatal("header invalid")
		}

		if len(row) < 1 {
			return
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			if err != nil {
				log.Fatal(err)
			}

			user := &User{
				Fullname: strings.TrimSpace(row[*head.FullnameIndex]),
				Username: strings.TrimSpace(row[*head.UsernameIndex]),
			}

			if pwd, err := util.HashPassword(strings.TrimSpace(row[*head.PasswordIndex])); err == nil {
				user.Password = pwd
				tx.Save(user)
			}
		}()
	}
}

func SeedQuestionExcel(rows *excelize.Rows) {
	head := &questionHeader{}

	wg := sync.WaitGroup{}
	defer wg.Wait()

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

		if len(row) < 1 {
			return
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			tx, err := DB().Begin(true)
			if err != nil {
				log.Fatal(err)
			}

			question := &Question{}

			num, err := strconv.ParseInt(strings.TrimSpace(row[*head.NumberIndex]), 10, 64)
			if err != nil {
				return
			}
			question.Number = int(num)
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
		}()
		rowIndex++
	}
	if err := rows.Close(); err != nil {
		fmt.Println(err)
	}
}

type questionHeader struct {
	NumberIndex   *int
	QuestionIndex *int

	// OptionsIndex contains indices of choices
	// The column looks like: Option 1 | Option 2 | Option A | Correct  etc.
	// So when iterating the row, and it's time for "Correct" column which contains like "1,2,A"
	// then we can split by comma(,) and check each correct options in OptionsIndex["1"] in current row.
	OptionsIndex *map[string]int
	CorrectIndex *int
}

func (h *questionHeader) Complete() bool {
	if h.NumberIndex == nil || h.QuestionIndex == nil || h.OptionsIndex == nil || h.CorrectIndex == nil {
		return false
	}

	return true
}

func (h *questionHeader) SetNumberIndex(i int) {
	if h.NumberIndex == nil {
		h.NumberIndex = new(int)
	}
	*h.NumberIndex = i
}

func (h *questionHeader) SetQuestionIndex(i int) {
	if h.QuestionIndex == nil {
		h.QuestionIndex = new(int)
	}
	*h.QuestionIndex = i
}

func (h *questionHeader) SetCorrectIndex(i int) {
	if h.CorrectIndex == nil {
		h.CorrectIndex = new(int)
	}
	*h.CorrectIndex = i
}

func (h *questionHeader) AddToOptions(value string, index int) {
	if h.OptionsIndex == nil {
		h.OptionsIndex = new(map[string]int)
		*h.OptionsIndex = make(map[string]int)
	}
	(*h.OptionsIndex)[value] = index
}

func (h *questionHeader) IsIndexInOptions(index int) bool {
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

func (h *questionHeader) GetOptions() map[string]int {
	return *h.OptionsIndex
}

type userHeader struct {
	FullnameIndex *int
	UsernameIndex *int
	PasswordIndex *int
}

func (uh *userHeader) Complete() bool {
	return uh.FullnameIndex != nil && uh.UsernameIndex != nil && uh.PasswordIndex != nil
}

func (uh *userHeader) SetFullnameIndex(i int) {
	if uh.FullnameIndex == nil {
		uh.FullnameIndex = new(int)
	}
	*uh.FullnameIndex = i
}

func (uh *userHeader) SetUsernameIndex(i int) {
	if uh.UsernameIndex == nil {
		uh.UsernameIndex = new(int)
	}
	*uh.UsernameIndex = i
}

func (uh *userHeader) SetPasswordIndex(i int) {
	if uh.PasswordIndex == nil {
		uh.PasswordIndex = new(int)
	}
	*uh.PasswordIndex = i
}
