package db

import (
	"log"
	"reflect"

	"github.com/94d/goquiz/db/query"
	"github.com/spf13/viper"
)

func Seed() {
	log.Println("Seeding...")

	usersConf := viper.New()
	usersConf.SetConfigName("users")
	usersConf.SetConfigType("yml")
	usersConf.AddConfigPath(".")

	err := usersConf.ReadInConfig()
	if err != nil {
		log.Println("Unable to seed,", err)
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Make sure to create users.yml; See users.example.yml for example")
		}
	}

	SeedUsers(usersConf.AllSettings())
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
