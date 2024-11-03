package todo

import (
	"context"
	"database/sql"
	"golang-template-htmx-alpine/gen/db"
	"golang-template-htmx-alpine/internal/model"
	"log/slog"
	"net/http"
)

type TodoService struct {
	queries *db.Queries
	logger  *slog.Logger
}

func New(logger *slog.Logger, queries *db.Queries) *TodoService {
	return &TodoService{
		queries: queries,
		logger:  logger,
	}
}

func (s *TodoService) List(ctx context.Context) ([]model.Todo, error) {
	todos, err := s.queries.GetTodos(ctx)
	if err != nil {
		s.logger.Error("error fetching todos", "error", err)
		return nil, err
	}

	var list []model.Todo
	for _, todo := range todos {
		list = append(list, model.Todo{
			Id:          todo.ID,
			Name:        todo.Name,
			Description: todo.Description.String,
		})
	}

	return list, nil
}

func (s *TodoService) FromRequest(r *http.Request) model.Todo {
	return model.Todo{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}
}

func (s *TodoService) CreateFromForm(ctx context.Context, t model.Todo) (model.Todo, error) {
	todo, err := s.queries.CreateTodo(ctx, db.CreateTodoParams{
		Name:        t.Name,
		Description: sql.NullString{String: t.Description, Valid: true},
	})

	if err != nil {
		s.logger.Error("error creating todo", "error", err)
		return model.Todo{}, err
	}

	return model.Todo{
		Id:          todo.ID,
		Name:        todo.Name,
		Description: todo.Description.String,
	}, nil
}

func (s *TodoService) FindById(ctx context.Context, id int64) (model.Todo, error) {
	todo, err := s.queries.GetTodo(ctx, id)
	if err != nil {
		s.logger.Error("error fetching todo", "error", err)
		return model.Todo{}, err
	}

	return model.Todo{
		Id:          todo.ID,
		Name:        todo.Name,
		Description: todo.Description.String,
	}, nil
}

func (s *TodoService) DeleteById(ctx context.Context, id int64) error {
	if err := s.queries.DeleteTodo(ctx, id); err != nil {
		s.logger.Error("error deleting todo", "error", err)
		return err
	}

	return nil
}

func (s *TodoService) Create(ctx context.Context, todo model.Todo) error {
	if _, err := s.queries.CreateTodo(ctx, db.CreateTodoParams{
		Name:        todo.Name,
		Description: sql.NullString{String: todo.Description, Valid: true},
	}); err != nil {
		s.logger.Error("error creating todo", "error", err)
		return err
	}

	return nil
}

func (s *TodoService) UpdateById(ctx context.Context, id int64, todo model.Todo) (model.Todo, error) {
	t, err := s.queries.UpdateTodo(ctx, db.UpdateTodoParams{
		ID:          id,
		Name:        todo.Name,
		Description: sql.NullString{String: todo.Description, Valid: true},
	})

	if err != nil {
		s.logger.Error("error updating todo", "error", err)
		return model.Todo{}, err
	}

	return model.Todo{
		Name:        t.Name,
		Description: t.Description.String,
	}, nil
}
