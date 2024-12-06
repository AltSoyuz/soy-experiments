package server

import (
	"net/http"

	"github.com/AltSoyuz/soy-experiments/apps/todo/auth"
	"github.com/AltSoyuz/soy-experiments/apps/todo/config"
	"github.com/AltSoyuz/soy-experiments/apps/todo/handlers"
	"github.com/AltSoyuz/soy-experiments/apps/todo/todo"
	"github.com/AltSoyuz/soy-experiments/lib/httpserver"
)

func New(
	config *config.Config,
	csrf *httpserver.CSRFProtection,
	authService *auth.Service,
	todoStore *todo.TodoStore,
) http.Handler {
	mux := http.NewServeMux()

	handlers.AddRoutes(
		config,
		csrf,
		mux,
		authService,
		todoStore,
	)

	return csrf.Middleware(mux)
}
