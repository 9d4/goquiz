package db

import (
	"log"
	"reflect"

	"github.com/94d/goquiz/db/query"
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

		_, err := query.InsertUser(DB(), &query.User{Fullname: usr["fullname"].(string), Username: usr["username"].(string), Password: usr["password"].(string)})
		if err != nil {
			log.Printf("username: %#v; err:%s", usr["username"], err)
			continue
		}
	}
}

func SeedQuestions(data map[string]interface{}) {
	_, ok := data["name"].(string)
	if !ok {
		log.Printf("Invalid quiz name, should be string; got %#v", data["name"])
		return
	}

	ref := reflect.TypeOf(data["questions"])
	if ref.Kind() != reflect.Slice {
		log.Printf("Invalid questions, should be array of question")
		return
	}

	questions := data["questions"].([]interface{})
	for _, question := range questions {
		q, ok := question.(map[string]interface{})
		if !ok {
			continue
		}

		body, ok := q["body"].(string)
		if !ok {
			continue
		}
		res, err := query.InsertQuestion(DB(), &query.Question{Body: body})
		if err != nil {
			continue
		}
		insertedID, err := res.LastInsertId()
		if err != nil {
			continue
		}

		rollback := func() {
			query.DeleteQuestionByID(DB(), int(insertedID))
		}

		if reflect.TypeOf(q["choices"]).Kind() != reflect.Slice {
			rollback()
			continue
		}

		choices, ok := q["choices"].([]interface{})
		if !ok {
			rollback()
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

			query.InsertChoice(DB(), query.NewChoice(body, int(insertedID), correct))
		}
	}
}
