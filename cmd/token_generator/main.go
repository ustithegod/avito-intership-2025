package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// makeToken предоставляет удобную генерацию валидных JWT.
func makeToken(secret string, role string) string {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": role,
		"exp":  time.Now().Add(365 * 24 * time.Hour).Unix(),
	})

	s, err := t.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	return s
}

func main() {
	adminSecret := "admin_secret_key"
	userSecret := "user_secret_key"

	fmt.Println("ADMIN_TOKEN=" + makeToken(adminSecret, "admin"))
	fmt.Println("USER_TOKEN=" + makeToken(userSecret, "user"))
}
