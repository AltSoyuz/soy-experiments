package todo

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/AltSoyuz/soy-experiments/apps/todo/gen/db"
	"github.com/AltSoyuz/soy-experiments/apps/todo/model"
	"github.com/AltSoyuz/soy-experiments/apps/todo/web/forms"
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
			IsComplete:  todo.IsComplete != 0,
		})
	}

	return list, nil
}

func (s *TodoStore) CreateFromForm(ctx context.Context, todoForm forms.TodoForm, userId int64) (model.Todo, error) {
	todo, err := s.queries.CreateTodo(ctx, db.CreateTodoParams{
		UserID:      userId,
		Name:        todoForm.Name,
		Description: sql.NullString{String: todoForm.Description, Valid: true},
	})

	if err != nil {
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
		IsComplete:  todo.IsComplete != 0,
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
	isComplete := int64(0)
	if todo.IsComplete {
		isComplete = 1
	}
	t, err := s.queries.UpdateTodo(ctx, db.UpdateTodoParams{
		ID:          todo.Id,
		UserID:      todo.UserId,
		Name:        todo.Name,
		Description: sql.NullString{String: todo.Description, Valid: true},
		IsComplete:  isComplete,
	})

	if err != nil {
		slog.Error("error updating todo", "error", err)
		return model.Todo{}, err
	}

	return model.Todo{
		Id:          t.ID,
		Name:        t.Name,
		Description: t.Description.String,
		UserId:      t.UserID,
		IsComplete:  t.IsComplete != 0,
	}, nil
}
