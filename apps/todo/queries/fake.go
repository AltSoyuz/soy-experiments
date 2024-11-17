package queries

import (
	"context"
	"crypto/rand"
	"errors"
	"golang-template-htmx-alpine/apps/todo/gen/db"
)

type FakeQuerier struct {
	Sessions map[string]db.CreateSessionParams
	Todos    map[int64]db.CreateTodoParams
}

func NewFakeQuerier() *FakeQuerier {
	return &FakeQuerier{
		Sessions: make(map[string]db.CreateSessionParams),
		Todos:    make(map[int64]db.CreateTodoParams),
	}
}

func (f *FakeQuerier) CreateSession(ctx context.Context, arg db.CreateSessionParams) (db.Session, error) {
	if arg.ID == "" || arg.UserID <= 0 {
		return db.Session{}, errors.New("invalid session parameters")
	}
	f.Sessions[arg.ID] = arg
	return db.Session{
		ID:        arg.ID,
		UserID:    arg.UserID,
		ExpiresAt: arg.ExpiresAt,
	}, nil
}

// Implement other methods as no-op or panics if not needed for this test
func (f *FakeQuerier) CreateTodo(ctx context.Context, arg db.CreateTodoParams) (db.Todo, error) {
	if arg.Name == "" {
		return db.Todo{}, errors.New("invalid todo parameters")
	}
	randomInt := make([]byte, 8)
	if _, err := rand.Read(randomInt); err != nil {
		return db.Todo{}, err
	}
	todo := db.Todo{
		ID:          int64(randomInt[0]),
		Name:        arg.Name,
		Description: arg.Description,
	}

	f.Todos[todo.ID] = arg

	return todo, nil
}

func (f *FakeQuerier) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	panic("not implemented")
}

func (f *FakeQuerier) DeleteSession(ctx context.Context, id string) error {
	if _, exists := f.Sessions[id]; !exists {
		return errors.New("session not found")
	}
	delete(f.Sessions, id)
	return nil
}

func (f *FakeQuerier) DeleteTodo(ctx context.Context, id int64) error {
	if _, exists := f.Todos[id]; !exists {
		return errors.New("todo not found")
	}
	delete(f.Todos, id)
	return nil
}

func (f *FakeQuerier) GetTodo(ctx context.Context, id int64) (db.Todo, error) {
	panic("not implemented")
}

func (f *FakeQuerier) GetTodos(ctx context.Context) ([]db.Todo, error) {
	if len(f.Todos) == 0 {
		return nil, nil
	}
	todos := make([]db.Todo, 0, len(f.Todos))
	for _, todo := range f.Todos {
		todos = append(todos, db.Todo{
			Name:        todo.Name,
			Description: todo.Description,
		})
	}
	return todos, nil
}

func (f *FakeQuerier) GetUserByUsername(ctx context.Context, username string) (db.User, error) {
	panic("not implemented")
}

func (f *FakeQuerier) UpdateSession(ctx context.Context, arg db.UpdateSessionParams) (db.Session, error) {
	if arg.ID == "" {
		return db.Session{}, errors.New("invalid session parameters")
	}
	session, exists := f.Sessions[arg.ID]
	if !exists {
		return db.Session{}, errors.New("session not found")
	}
	session.ExpiresAt = arg.ExpiresAt
	f.Sessions[arg.ID] = session
	return db.Session{
		ID:        session.ID,
		UserID:    session.UserID,
		ExpiresAt: session.ExpiresAt,
	}, nil

}

func (f *FakeQuerier) UpdateTodo(ctx context.Context, arg db.UpdateTodoParams) (db.Todo, error) {
	if arg.ID == 0 {
		return db.Todo{}, errors.New("invalid todo parameters")
	}
	todo, exists := f.Todos[arg.ID]
	if !exists {
		return db.Todo{}, errors.New("todo not found")
	}
	todo.Name = arg.Name
	todo.Description = arg.Description
	f.Todos[arg.ID] = todo
	return db.Todo{
		Name:        todo.Name,
		Description: todo.Description,
	}, nil
}

func (f *FakeQuerier) ValidateSessionToken(ctx context.Context, id string) (db.ValidateSessionTokenRow, error) {
	session, exists := f.Sessions[id]
	if !exists {
		return db.ValidateSessionTokenRow{}, errors.New("session not found")
	}
	return db.ValidateSessionTokenRow{
		ID:        session.ID,
		UserID:    session.UserID,
		ExpiresAt: session.ExpiresAt,
	}, nil
}
