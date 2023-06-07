package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken(t *testing.T) {
	type args struct {
		claims *ClaimsData
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Gen 1",
			args: args{claims: &ClaimsData{Fullname: "Foo Bar", Username: "foobar"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := GenerateToken(tt.args.claims, []byte("test")); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	type args struct {
		token   string
		keyFunc jwt.Keyfunc
	}
	tests := []struct {
		name      string
		args      args
		wantValid bool
		wantErr   bool
	}{
		{
			name: "Token Gen and Parse 1",
			args: args{
				token: (func() string {
					str, _ := GenerateToken(&ClaimsData{Fullname: "Foo Bar", Username: "foobar"}, []byte("test"))
					return str
				})(),
				keyFunc: func(t *jwt.Token) (interface{}, error) {
					return []byte("test"), nil
				},
			},
			wantValid: true,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid, err := ValidateToken(tt.args.token, tt.args.keyFunc)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValid != tt.wantValid {
				t.Errorf("ValidateToken() = %v, want %v", gotValid, tt.wantValid)
			}
		})
	}
}
