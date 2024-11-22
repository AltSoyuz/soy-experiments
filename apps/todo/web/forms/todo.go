package forms

import (
	"fmt"
	"net/http"
)

type TodoForm struct {
	Name        string `form:"name"`
	Description string `form:"description"`
}

func TodoFrom(r *http.Request) (TodoForm, error) {
	err := r.ParseForm()
	if err != nil {
		return TodoForm{}, err
	}
	name := r.FormValue("name")
	description := r.FormValue("description")

	if name == "" {
		return TodoForm{}, fmt.Errorf("name is required")
	}

	return TodoForm{
		Name:        name,
		Description: description,
	}, nil
}
