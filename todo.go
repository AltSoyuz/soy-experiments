package main

import "net/http"

var todos = []Todo{
	{
		Name:        "Todo 1",
		Description: "Description 1",
	},
}

func createTodoFromForm(r *http.Request) Todo {
	err := r.ParseForm()
	if err != nil {
		return Todo{}
	}
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
