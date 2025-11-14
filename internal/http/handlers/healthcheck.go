package handlers

import (
	"net/http"

	"github.com/go-chi/render"
)

func Healthcheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, map[string]string{"status": "ok"})
	}
}
