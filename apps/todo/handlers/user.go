package handlers

import (
	"golang-template-htmx-alpine/apps/todo/auth"
	"golang-template-htmx-alpine/apps/todo/gen/db"
	"golang-template-htmx-alpine/apps/todo/views"
	"golang-template-htmx-alpine/lib/argon2id"
	"log/slog"
	"net/http"
)

func CreateUser(render views.RenderFunc, queries *db.Queries) http.HandlerFunc {
	type Data struct {
		Message string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		if password == "" || len(password) > 127 {
			slog.Warn("invalid password", "password", password)
			w.WriteHeader(http.StatusBadRequest)
			render(w, Data{Message: "Invalid password"}, "error-msg")
			return
		}

		err := auth.VerifyPasswordStrength(password)
		if err != nil {
			slog.Warn("password does not meet requirements", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			render(w, Data{Message: "Password does not meet requirements"}, "error-msg")
			return
		}

		passwordHash, err := argon2id.Hash(password)
		if err != nil {
			slog.Error("error hashing password", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = queries.CreateUser(r.Context(), db.CreateUserParams{
			Username:     username,
			PasswordHash: passwordHash,
		})

		if err != nil {
			slog.Error("error creating user", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusNoContent)
		return
	}
}
