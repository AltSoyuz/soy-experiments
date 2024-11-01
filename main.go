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
		Name:        "Todo1",
		Description: "Description 1",
	},
}

type TodoPageData struct {
	Title string
	Items []Todo
}

type EditTodoPageData struct {
	Title string
	Item  Todo
}

func main() {
	http.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("todos.html")
		page := TodoPageData{
			Title: "My Todo List",
			Items: todos,
		}
		t.Execute(w, page)
	})

	http.HandleFunc("/edit/", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("edit.html")
		name := r.URL.Path[len("/edit/"):]
		page := EditTodoPageData{
			Title: name,
		}

		for i := 0; i < len(todos); i++ {
			if todos[i].Name == name {
				page.Item.Name = todos[i].Name
				page.Item.Description = todos[i].Description
			}
		}

		t.Execute(w, page)
	})

	http.HandleFunc("/save/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/save/"):]
		todoName := r.FormValue("name")
		descriptionName := r.FormValue("description")

		if name == "" {
			todos = append(todos, Todo{
				Name:        todoName,
				Description: descriptionName,
			})
		}

		for i := 0; i < len(todos); i++ {
			if todos[i].Name == name {
				todos[i].Name = todoName
				todos[i].Description = descriptionName
			}
		}

		http.Redirect(w, r, "/todos/", http.StatusSeeOther)
	})

	http.HandleFunc("/new/", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("new.html")
		t.Execute(w, nil)
	})

	http.HandleFunc("/delete/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/delete/"):]

		for i := 0; i < len(todos); i++ {
			if todos[i].Name == name {
				todos = append(todos[:i], todos[i+1:]...)
				break
			}
		}

		http.Redirect(w, r, "/todos/", http.StatusSeeOther)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
