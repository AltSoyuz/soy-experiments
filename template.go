package main

import (
	"embed"
	"html/template"
	"io"
	"log/slog"
)

//go:embed index.html todo.html form.html
var todoTemplate embed.FS

// initialize the template
func InitTodoTemplates() (*template.Template, error) {
	tmpl, err := template.ParseFS(todoTemplate, "index.html", "todo.html", "form.html")
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

type renderFunc func(w io.Writer, data interface{}, name string)

func newRender(logger *slog.Logger, tmpl *template.Template) renderFunc {
	return func(w io.Writer, data interface{}, name string) {
		err := tmpl.ExecuteTemplate(w, name, data)
		if err != nil {
			logger.Error("failed to execute template", "error", err, "template", name)
		}
	}
}
