package main

import (
	"embed"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
)

//go:embed dist/*
var files embed.FS

func main() {
	handler := http.NewServeMux()

	distDir, err := fs.Sub(files, "dist")
	if err != nil {
		log.Fatalf("error reading embeded files: %v", err)
	}
	fileServer := http.FileServer(http.FS(distDir))

	handler.Handle("/", http.StripPrefix("/", fileServer))
	handler.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	addRoutes(handler)

	slog.Info("starting http server on 3000")
	log.Fatal(http.ListenAndServe(":3000", handler))
}

func addRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/notes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"text":"Buy milk"},{"id":2,"text":"Call mom"}]`))
	})
}
