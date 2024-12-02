package forms

import (
	"fmt"
	"net/http"
)

type RegisterForm struct {
	Email           string `form:"email"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm-password"`
}

func RegisterFrom(r *http.Request) (RegisterForm, error) {
	err := r.ParseForm()
	if err != nil {
		return RegisterForm{}, err
	}

	form := RegisterForm{
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirm-password"),
	}

	if form.Email == "" {
		return RegisterForm{}, fmt.Errorf("email is required")
	}
	if form.Password == "" {
		return RegisterForm{}, fmt.Errorf("password is required")
	}
	if form.ConfirmPassword == "" {
		return RegisterForm{}, fmt.Errorf("confirm password is required")
	}
	if form.Password != form.ConfirmPassword {
		return RegisterForm{}, fmt.Errorf("passwords do not match")
	}

	return form, nil
}

type LoginForm struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func LoginFrom(r *http.Request) (LoginForm, error) {
	err := r.ParseForm()

	if err != nil {
		return LoginForm{}, err
	}

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
