package main

import (
	"golang-template-htmx-alpine/apps/todo/views"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
	render views.RenderFunc,
	todoService *TodoService,
) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	mux.HandleFunc("GET /health", HealthHandler)
	mux.HandleFunc("GET /", todoListHandler(render, todoService))
	mux.HandleFunc("POST /todos", createTodoHandler(render, todoService))
	mux.HandleFunc("GET /todos/{id}/form", getTodoFormHandler(render, todoService))
	mux.HandleFunc("PUT /todos/{id}", updateTodoHandler(render, todoService))
	mux.HandleFunc("DELETE /todos/{id}", deleteTodoHandler(todoService))
}
