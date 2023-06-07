package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ClaimsData struct {
	Fullname string
	Username string
}

func GenerateToken(claims *ClaimsData, key []byte) (string, error) {
	token := GenerateTokenRaw(claims)

	return token.SignedString(key)
}

func GenerateTokenRaw(claims *ClaimsData) *jwt.Token {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"fullname": claims.Fullname,
		"username": claims.Username,
		"iat":      time.Now().Unix(),
	})

	return token
}

func ValidateToken(token string, keyFunc jwt.Keyfunc) (valid bool, err error) {
	t, err := ParseToken(token, keyFunc)
	if err != nil {
		return
	}

	valid = t.Valid
	return
}

func ParseToken(token string, keyFunc jwt.Keyfunc) (t *jwt.Token, err error) {
	t, err = jwt.ParseWithClaims(token, jwt.MapClaims{}, keyFunc)
	return
}

func KeyFunc(key []byte) jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		return key, nil
	}
}
