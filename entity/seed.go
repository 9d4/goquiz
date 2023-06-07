package entity

import (
	"log"
	"reflect"

	"github.com/spf13/viper"
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
