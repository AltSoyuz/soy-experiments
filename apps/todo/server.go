package main

import (
	"golang-template-htmx-alpine/apps/todo/views"
	"net/http"
)

func newServer(
	render views.RenderFunc,
	todoService *TodoService,
) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, render, todoService)

	var handler http.Handler = mux
	return handler
}
