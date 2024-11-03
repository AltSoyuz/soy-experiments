package templates

import (
	"embed"
	"html/template"
	"io"
	"log/slog"
)

//go:embed components/* pages/*
var embedded embed.FS

func New(logger *slog.Logger) (RenderFunc, error) {
	tmpl, err := template.ParseFS(embedded, "pages/*.html", "components/*.html")
	if err != nil {
		return nil, err
	}

	initializeTemplates("pages", "page", logger)
	initializeTemplates("components", "component", logger)

	render := newRender(logger, tmpl)

	return render, nil
}

func initializeTemplates(dir string, templateType string, logger *slog.Logger) {
	templates, err := embedded.ReadDir(dir)
	if err != nil || len(templates) == 0 {
		logger.Warn("no " + templateType + " templates found")
	} else {
		var templateNames []string
		for _, file := range templates {
			templateNames = append(templateNames, file.Name())
		}
		logger.Info(templateType+" templates initialized", templateType, templateNames)
	}
}

type RenderFunc func(w io.Writer, data interface{}, name string)

func newRender(logger *slog.Logger, tmpl *template.Template) RenderFunc {
	return func(w io.Writer, data interface{}, name string) {
		err := tmpl.ExecuteTemplate(w, name, data)
		if err != nil {
			logger.Error("failed to execute template", "error", err, "template", name)
		}
	}
}
