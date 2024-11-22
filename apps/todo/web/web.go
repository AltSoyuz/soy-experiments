package web

import (
	"embed"
	"html/template"
	"io"
	"log/slog"
)

type PageData struct {
	Title string
}

type Renderer struct {
	components *template.Template
	pages      *template.Template
}

//go:embed components/* pages/*
var embedded embed.FS

func NewRender(components *template.Template, pages *template.Template) *Renderer {
	return &Renderer{
		components: components,
		pages:      pages,
	}
}

func Init() (*Renderer, error) {
	componentTemplates, err := template.ParseFS(embedded, "components/*.html")
	if err != nil {
		return nil, err
	}
	initializeTemplates("components", "component")

	pageTemplates, err := template.ParseFS(embedded, "pages/*.html")
	if err != nil {
		return nil, err
	}
	initializeTemplates("pages", "page")

	render := NewRender(componentTemplates, pageTemplates)

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

func (r *Renderer) RenderPage(w io.Writer, data interface{}, name string) {
	// check if template exists
	if r.pages.Lookup(name) == nil {
		slog.Error("template not found", "template", name)
		return
	}
	err := r.pages.ExecuteTemplate(w, name, data)
	if err != nil {
		slog.Error("failed to execute template", "error", err, "template", name)
	}
}

func (r *Renderer) RenderComponent(w io.Writer, data interface{}, name string) {
	// check if template exists
	if r.components.Lookup(name) == nil {
		slog.Error("template not found", "template", name)
		return
	}
	err := r.components.ExecuteTemplate(w, name, data)
	if err != nil {
		slog.Error("failed to execute template", "error", err, "template", name)
	}
}

type ErrorFragmentData struct {
	Message string
}

func (r *Renderer) ErrorFragment(w io.Writer, message string) {
	r.RenderComponent(w, ErrorFragmentData{Message: message}, "error-msg")
}
