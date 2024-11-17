package views

import (
	"embed"
	"html/template"
	"io"
	"log/slog"
)

//go:embed components/* pages/*
var embedded embed.FS

func Init() (RenderFunc, error) {
	tmpl, err := template.ParseFS(embedded, "pages/*.html", "components/*.html")
	if err != nil {
		return nil, err
	}

	initializeTemplates("pages", "page")
	initializeTemplates("components", "component")

	render := newRender(tmpl)

	return render, nil
}

func initializeTemplates(dir string, templateType string) {
	templates, err := embedded.ReadDir(dir)
	if err != nil || len(templates) == 0 {
		slog.Warn("no " + templateType + " templates found")
	} else {
		var templateNames []string
		for _, file := range templates {
			templateNames = append(templateNames, file.Name())
		}
		slog.Info(templateType+" templates initialized", templateType, templateNames)
	}
}

type RenderFunc func(w io.Writer, data interface{}, name string)

func newRender(tmpl *template.Template) RenderFunc {
	return func(w io.Writer, data interface{}, name string) {
		// check if template exists
		if tmpl.Lookup(name) == nil {
			slog.Error("template not found", "template", name)
			return
		}
		err := tmpl.ExecuteTemplate(w, name, data)
		if err != nil {
			slog.Error("failed to execute template", "error", err, "template", name)
		}
	}
}
