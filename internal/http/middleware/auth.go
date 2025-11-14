package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"avito-intership-2025/internal/http/api"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
)

type key int

const RoleKey key = 1

func AuthMiddleware(next http.Handler) http.Handler {
	adminSecret := os.Getenv("ADMIN_JWT_SECRET")
	userSecret := os.Getenv("USER_JWT_SECRET")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		if tokenString == "" {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, api.Error(api.ErrCodeNotFound, "resource not found"))
			return
		}

		tokenString, _ = strings.CutPrefix(tokenString, "Bearer ")

		// Try admin token
		role, ok := validateToken(tokenString, adminSecret)
		if ok && role == "admin" {
			ctx := context.WithValue(r.Context(), RoleKey, "admin")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Try user token
		role, ok = validateToken(tokenString, userSecret)
		if ok && role == "user" {
			ctx := context.WithValue(r.Context(), RoleKey, "user")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		http.Error(w, "invalid token", http.StatusUnauthorized)
	})
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(RoleKey).(string)

		if role != "admin" {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, api.Error(api.ErrCodeNotFound, "resource not found"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func validateToken(tokenString, secret string) (string, bool) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return "", false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		roleVal, ok := claims["role"].(string)
		if !ok {
			return "", false
		}
		return roleVal, true
	}

	return "", false
}
