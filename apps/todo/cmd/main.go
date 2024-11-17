package main

import (
	"context"
	"fmt"
	"golang-template-htmx-alpine/apps/todo/server"
	"os"
)

func main() {
	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
