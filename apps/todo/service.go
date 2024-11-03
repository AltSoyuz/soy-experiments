package main

import (
	"context"
	"database/sql"
	"golang-template-htmx-alpine/apps/todo/gen/db"
	"log/slog"
	"net/http"
)

type TodoService struct {
	queries *db.Queries
}

func newTodoService(queries *db.Queries) *TodoService {
	return &TodoService{
		queries: queries,
	}
}

func (s *TodoService) List(ctx context.Context) ([]Todo, error) {
	todos, err := s.queries.GetTodos(ctx)
	if err != nil {
		slog.Error("error fetching todos", "error", err)
		return nil, err
	}

	var list []Todo
	for _, todo := range todos {
		list = append(list, Todo{
			Id:          todo.ID,
			Name:        todo.Name,
			Description: todo.Description.String,
		})
	}

	return list, nil
}

func (s *TodoService) FromRequest(r *http.Request) Todo {
	return Todo{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}
}

func (s *TodoService) CreateFromForm(ctx context.Context, t Todo) (Todo, error) {
	todo, err := s.queries.CreateTodo(ctx, db.CreateTodoParams{
		Name:        t.Name,
		Description: sql.NullString{String: t.Description, Valid: true},
	})

	if err != nil {
		slog.Error("error creating todo", "error", err)
		return Todo{}, err
	}

	return Todo{
		Id:          todo.ID,
		Name:        todo.Name,
		Description: todo.Description.String,
	}, nil
}

func (s *TodoService) FindById(ctx context.Context, id int64) (Todo, error) {
	todo, err := s.queries.GetTodo(ctx, id)
	if err != nil {
		slog.Error("error fetching todo", "error", err)
		return Todo{}, err
	}

	return Todo{
		Id:          todo.ID,
		Name:        todo.Name,
		Description: todo.Description.String,
	}, nil
}

func (s *TodoService) DeleteById(ctx context.Context, id int64) error {
	if err := s.queries.DeleteTodo(ctx, id); err != nil {
		slog.Error("error deleting todo", "error", err)
		return err
	}

	return nil
}

func (s *TodoService) Create(ctx context.Context, todo Todo) error {
	if _, err := s.queries.CreateTodo(ctx, db.CreateTodoParams{
		Name:        todo.Name,
		Description: sql.NullString{String: todo.Description, Valid: true},
	}); err != nil {
		slog.Error("error creating todo", "error", err)
		return err
	}

	return nil
}

func (s *TodoService) UpdateById(ctx context.Context, id int64, todo Todo) (Todo, error) {
	t, err := s.queries.UpdateTodo(ctx, db.UpdateTodoParams{
		ID:          id,
		Name:        todo.Name,
		Description: sql.NullString{String: todo.Description, Valid: true},
	})

	if err != nil {
		slog.Error("error updating todo", "error", err)
		return Todo{}, err
	}

	return Todo{
		Name:        t.Name,
		Description: t.Description.String,
	}, nil
}
