package todo

import (
	"golang-template-htmx-alpine/internal/model"
	"golang-template-htmx-alpine/internal/store"
	"net/http"
)

func CreateFromForm(r *http.Request) model.Todo {
	err := r.ParseForm()
	if err != nil {
		return model.Todo{}
	}
	return model.Todo{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}
}

func FindByName(name string) model.Todo {
	for i := 0; i < len(store.Data); i++ {
		if store.Data[i].Name == name {
			return store.Data[i]
		}
	}
	return model.Todo{}
}

func DeleteByName(name string) {
	for i := 0; i < len(store.Data); i++ {
		if store.Data[i].Name == name {
			store.Data = append(store.Data[:i], store.Data[i+1:]...)
		}
	}
}

func Add(todo model.Todo) {
	store.Data = append(store.Data, todo)
}

func UpdateByName(name string, todo model.Todo) {
	for i := 0; i < len(store.Data); i++ {
		if store.Data[i].Name == name {
			store.Data[i] = todo
		}
	}
}
