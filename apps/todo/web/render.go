package web

import (
	"golang-template-htmx-alpine/apps/todo/model"
	"io"
)

type loginPageData struct {
	Title string
}

func RenderRegister(w io.Writer) {
	RenderPage(w, "register", loginPageData{Title: "Register"})
}

func RenderLogin(w io.Writer) {
	RenderPage(w, "login", loginPageData{Title: "Login"})
}

func RenderVerifyEmail(w io.Writer) {
	RenderPage(w, "verify-email", loginPageData{Title: "Verify Email"})
}

func RenderAbout(w io.Writer) {
	RenderPage(w, "about", loginPageData{Title: "About"})
}

func Render404(w io.Writer) {
	RenderPage(w, "404", loginPageData{Title: "404"})
}

type TodoPageData struct {
	Title string
	Items []model.Todo
	Email string
}

func RenderTodoList(w io.Writer, page TodoPageData) {
	RenderPage(w, "index", page)
}

func RenderFormFragment(w io.Writer, todo model.Todo) {
	RenderComponent(w, "form", "form", todo)
}

func RenderTodoFragment(w io.Writer, todo model.Todo) {
	RenderComponent(w, "todo", "todo", todo)
}

func RenderError(w io.Writer, message string) {
	RenderComponent(w, "error-msg", "error-msg", map[string]interface{}{
		"Message": message,
	})
}
