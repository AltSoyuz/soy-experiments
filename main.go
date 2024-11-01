package main

import (
	"html/template"
	"log"
	"net/http"
)

type Todo struct {
	Name        string
	Description string
}

var todos = []Todo{
	{
		Name:        "Todo 1",
		Description: "Description 1",
	},
}

var t *template.Template

type TodoPageData struct {
	Title string
	Items []Todo
}

type EditTodoPageData struct {
	Title string
	Item  Todo
}

func createTodoFromForm(r *http.Request) Todo {
	return Todo{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}
}

func findTodoByName(name string) Todo {
	for i := 0; i < len(todos); i++ {
		if todos[i].Name == name {
			return todos[i]
		}
	}
	return Todo{}
}

func deleteTodoByName(name string) {
	for i := 0; i < len(todos); i++ {
		if todos[i].Name == name {
			todos = append(todos[:i], todos[i+1:]...)
		}
	}
}

func addTodo(todo Todo) {
	todos = append(todos, todo)
}

func updateTodoByName(name string, todo Todo) {
	for i := 0; i < len(todos); i++ {
		if todos[i].Name == name {
			todos[i] = todo
		}
	}
}

func main() {
	t, _ = template.ParseFiles("index.html", "todo.html", "form.html")

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		page := TodoPageData{
			Title: "My Todo List",
			Items: todos,
		}
		t.ExecuteTemplate(w, "index.html", page)
	})

	http.HandleFunc("POST /todos", func(w http.ResponseWriter, r *http.Request) {
		todo := createTodoFromForm(r)

		addTodo(todo)

		t.ExecuteTemplate(w, "todo", todo)
	})

	http.HandleFunc("GET /todos/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")

		todo := findTodoByName(name)

		t.ExecuteTemplate(w, "form", todo)
	})

	http.HandleFunc("PUT /todos/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		todo := createTodoFromForm(r)

		updateTodoByName(name, todo)

		t.ExecuteTemplate(w, "todo", todo)
	})

	http.HandleFunc("DELETE /todos/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")

		deleteTodoByName(name)

		w.WriteHeader(http.StatusOK)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
