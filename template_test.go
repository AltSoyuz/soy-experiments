package main

import (
	"bytes"
	"io"
	"log/slog"
	"testing"
)

func TestInitTodoTemplates(t *testing.T) {
	tmpl, err := InitTodoTemplates()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tmpl == nil {
		t.Fatalf("expected template, got nil")
	}
}

func TestNewRender(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tmpl, err := InitTodoTemplates()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	render := newRender(logger, tmpl)
	var buf bytes.Buffer
	data := map[string]string{"Title": "Test Title"}

	render(&buf, data, "index.html")

	if buf.Len() == 0 {
		t.Fatalf("expected non-empty buffer, got empty buffer")
	}
}
