package auth

import (
	"crypto/rand"
	"fmt"
	"log"
)

func GenerateSecret() string {
	secret := make([]byte, 16)
	_, err := rand.Read(secret)
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", secret)
}
