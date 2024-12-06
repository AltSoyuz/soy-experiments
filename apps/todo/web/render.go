package web

import (
	"golang-template-htmx-alpine/apps/todo/model"
	"io"
)

type pageData struct {
	Title     string
	CSRFToken string
	Error     string
}

func RenderNotFoundPage(w io.Writer) {
	RenderPage(w, "404", nil)
}

func RenderRegisterPage(w io.Writer, csrfToken string) {
	RenderPage(w, "register", pageData{Title: "Register", CSRFToken: csrfToken})
}

func RenderLoginPage(w io.Writer, csrfToken string) {
	RenderPage(w, "login", pageData{Title: "Login", CSRFToken: csrfToken})
}

type FormData struct {
	CSRFToken string
	Error     string
}

func RenderLoginForm(w io.Writer, csrfToken, error string) {
	RenderComponent(w, "login-form", "login-form", FormData{CSRFToken: csrfToken, Error: error})
}

func RenderRegisterForm(w io.Writer, csrfToken, error string) {
	RenderComponent(w, "register-form", "register-form", FormData{CSRFToken: csrfToken, Error: error})
}

func RenderVerifyEmailForm(w io.Writer, csrfToken, error string) {
	RenderComponent(w, "verify-email-form", "verify-email-form", FormData{CSRFToken: csrfToken, Error: error})
}

func RenderVerifyEmail(w io.Writer, csrfToken string) {
	RenderPage(
		w,
		"verify-email",
		pageData{Title: "Verify Email", CSRFToken: csrfToken},
	)
}

func RenderAbout(w io.Writer) {
	RenderPage(w, "about", pageData{Title: "About"})
}

func Render404(w io.Writer) {
	RenderPage(w, "404", pageData{Title: "404"})
}

type TodoPageData struct {
	Title     string
	Items     []TodoComponentData
	Email     string
	CSRFToken string
}

func RenderTodoList(w io.Writer, page TodoPageData) {
	RenderPage(w, "index", page)
}

type TodoComponentData struct {
	Todo      model.Todo
	CSRFToken string
}

func RenderFormFragment(w io.Writer, todo model.Todo, csrfToken string) {
	RenderComponent(w, "todo-form", "todo-form", TodoComponentData{Todo: todo, CSRFToken: csrfToken})
}

func RenderTodoFragment(w io.Writer, todo model.Todo, csrfToken string) {
	RenderComponent(w, "todo", "todo", TodoComponentData{Todo: todo, CSRFToken: csrfToken})
}
