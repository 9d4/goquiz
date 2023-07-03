package auth

import (
	"crypto/rand"
	"fmt"

	"github.com/94d/goquiz/util"
)

func GenerateSecret() string {
	secret := make([]byte, 16)
	_, err := rand.Read(secret)
	if err != nil {
		util.Fatal(err)
	}

	return fmt.Sprintf("%x", secret)
}
