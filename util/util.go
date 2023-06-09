package util

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func Shuffle[T any](slice []T) []T {
	rand.Seed(time.Now().UnixNano())
	cursor := len(slice)

	for cursor > 0 {
		index := rand.Intn(cursor)
		cursor--

		temp := slice[cursor]
		slice[cursor] = slice[index]
		slice[index] = temp
	}

	return slice
}

func Fatal(v ...any) {
	log.Println(v...)
	fmt.Print("Something went wrong, read the error message. Press enter to exit")
	fmt.Scanln()
	os.Exit(1)
}
