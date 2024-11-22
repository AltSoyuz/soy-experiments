package forms

import (
	"fmt"
	"net/http"
)

type LoginForm struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func LoginFrom(r *http.Request) (LoginForm, error) {
	r.ParseForm()

	form := LoginForm{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if form.Email == "" {
		return LoginForm{}, fmt.Errorf("email is required")
	}
	if form.Password == "" {
		return LoginForm{}, fmt.Errorf("password is required")
	}

	return form, nil
}
