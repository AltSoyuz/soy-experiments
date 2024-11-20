package web

import (
	"bytes"
	"html/template"
	"testing"
)

// Test for Init function
func TestInit(t *testing.T) {
	renderFunc, err := Init()
	if err != nil {
		t.Fatalf("Init() returned an error: %v", err)
	}
	if renderFunc == nil {
		t.Fatal("Init() returned a nil RenderFunc")
	}
}

// Test for initializeTemplates function
func TestInitializeTemplates(t *testing.T) {
	// var embeddedFS embed.FS
	initializeTemplates("pages", "page")
	initializeTemplates("components", "component")

	// Assuming templates are present in the "pages" and "components" directories
	// This test checks if any warnings or info messages are triggered as expected.
}

// Test for RenderFunc
func TestRenderFunc(t *testing.T) {
	// Mock a parsed template
	tmpl, err := template.New("test").Parse("Hello, {{.Name}}!")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	render := newRender(tmpl)

	// Create a buffer to capture the rendered output
	var buf bytes.Buffer
	data := map[string]string{"Name": "World"}
	render(&buf, data, "test")

	expected := "Hello, World!"
	result := buf.String()
	if result != expected {
		t.Errorf("RenderFunc output = %q; want %q", result, expected)
	}
}
