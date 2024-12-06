package main

import (
	"context"
	"fmt"
	"os"

	"github.com/AltSoyuz/soy-experiments/apps/todo/server"
)

func main() {
	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
