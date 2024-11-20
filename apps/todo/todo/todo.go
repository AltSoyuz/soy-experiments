package todo

import (
	"context"
	"database/sql"
	"golang-template-htmx-alpine/apps/todo/gen/db"
	"golang-template-htmx-alpine/apps/todo/model"
	"log/slog"
)

type TodoStore struct {
	queries db.Querier
}

func Init(queries db.Querier) *TodoStore {
	return &TodoStore{
		queries: queries,
	}
}

func (s *TodoStore) List(ctx context.Context, userId int64) ([]model.Todo, error) {
	todos, err := s.queries.GetTodos(ctx, userId)
	if err != nil {
		slog.Error("error fetching todos", "error", err)
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

func (s *TodoStore) CreateFromForm(ctx context.Context, t model.Todo, userId int64) (model.Todo, error) {
	todo, err := s.queries.CreateTodo(ctx, db.CreateTodoParams{
		UserID:      userId,
		Name:        t.Name,
		Description: sql.NullString{String: t.Description, Valid: true},
	})

	if err != nil {
		slog.Error("error creating todo", "error", err)
		return model.Todo{}, err
	}

	return model.Todo{
		Id:          todo.ID,
		Name:        todo.Name,
		Description: todo.Description.String,
		UserId:      userId,
	}, nil
}

func (s *TodoStore) FindById(ctx context.Context, id, userId int64) (model.Todo, error) {
	todo, err := s.queries.GetTodo(ctx, db.GetTodoParams{
		ID:     id,
		UserID: userId,
	})
	if err != nil {
		slog.Error("error fetching todo", "error", err)
		return model.Todo{}, err
	}

	return model.Todo{
		Id:          todo.ID,
		Name:        todo.Name,
		UserId:      todo.UserID,
		Description: todo.Description.String,
	}, nil
}

func (s *TodoStore) Delete(ctx context.Context, id, userId int64) error {
	if err := s.queries.DeleteTodo(ctx, db.DeleteTodoParams{
		ID:     id,
		UserID: userId,
	}); err != nil {
		slog.Error("error deleting todo", "error", err)
		return err
	}

	return nil
}

func (s *TodoStore) Create(ctx context.Context, todo model.Todo) error {
	if _, err := s.queries.CreateTodo(ctx, db.CreateTodoParams{
		Name:        todo.Name,
		UserID:      todo.UserId,
		Description: sql.NullString{String: todo.Description, Valid: true},
	}); err != nil {
		slog.Error("error creating todo", "error", err)
		return err
	}

	return nil
}

func (s *TodoStore) Update(ctx context.Context, todo model.Todo) (model.Todo, error) {
	t, err := s.queries.UpdateTodo(ctx, db.UpdateTodoParams{
		ID:          todo.Id,
		UserID:      todo.UserId,
		Name:        todo.Name,
		Description: sql.NullString{String: todo.Description, Valid: true},
	})

	if err != nil {
		slog.Error("error updating todo", "error", err)
		return model.Todo{}, err
	}

	return model.Todo{
		Name:        t.Name,
		Description: t.Description.String,
		UserId:      t.UserID,
	}, nil
}
