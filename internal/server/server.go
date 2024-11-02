package server

import (
	"golang-template-htmx-alpine/internal/templates"
	"log/slog"
	"net/http"
)

func New(logger *slog.Logger, render templates.RenderFunc) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, logger, render)
	var handler http.Handler = mux

	return handler
}
