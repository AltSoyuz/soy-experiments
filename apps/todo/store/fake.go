package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"golang-template-htmx-alpine/apps/todo/gen/db"
	"time"
)

type FakeQuerier struct {
	Sessions                  map[string]db.CreateSessionParams
	Todos                     map[int64]db.CreateTodoParams
	Users                     map[int64]db.CreateUserParams
	EmailVerificationRequests map[int64]db.InsertUserEmailVerificationRequestParams
}

func NewFakeQuerier() *FakeQuerier {
	return &FakeQuerier{
		Sessions:                  make(map[string]db.CreateSessionParams),
		Todos:                     make(map[int64]db.CreateTodoParams),
		Users:                     make(map[int64]db.CreateUserParams),
		EmailVerificationRequests: make(map[int64]db.InsertUserEmailVerificationRequestParams),
	}
}

func (f *FakeQuerier) CreateSession(ctx context.Context, arg db.CreateSessionParams) (db.Session, error) {
	if arg.ID == "" || arg.UserID <= 0 {
		return db.Session{}, errors.New("invalid session parameters")
	}
	f.Sessions[arg.ID] = arg
	return db.Session(db.Session{
		ID:     arg.ID,
		UserID: arg.UserID,
	}), nil
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
		UserID:      arg.UserID,
	}

	f.Todos[todo.ID] = arg

	return todo, nil
}

func (f *FakeQuerier) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	if arg.Email == "" || arg.PasswordHash == "" {
		return db.User{}, errors.New("invalid user parameters")
	}
	randomInt := make([]byte, 8)
	if _, err := rand.Read(randomInt); err != nil {
		return db.User{}, err
	}
	user := db.User{
		ID:           int64(randomInt[0]),
		Email:        arg.Email,
		PasswordHash: arg.PasswordHash,
	}
	f.Users[user.ID] = arg
	return user, nil
}

func (f *FakeQuerier) DeleteSession(ctx context.Context, id string) error {
	if _, exists := f.Sessions[id]; !exists {
		return errors.New("session not found")
	}
	delete(f.Sessions, id)
	return nil
}

func (f *FakeQuerier) DeleteTodo(ctx context.Context, arg db.DeleteTodoParams) error {
	if _, exists := f.Todos[arg.ID]; !exists {
		return errors.New("todo not found")
	}

	for _, todo := range f.Todos {
		if todo.UserID != arg.UserID {
			return errors.New("unauthorized")
		}
	}

	delete(f.Todos, arg.ID)
	return nil
}

func (f *FakeQuerier) GetTodo(ctx context.Context, arg db.GetTodoParams) (db.Todo, error) {
	panic("not implemented")
}

func (f *FakeQuerier) GetTodos(ctx context.Context, userId int64) ([]db.Todo, error) {
	for _, todo := range f.Todos {
		if todo.UserID == userId {
			return []db.Todo{
				{
					Name:        todo.Name,
					Description: todo.Description,
					UserID:      todo.UserID,
				},
			}, nil
		}
	}
	return nil, errors.New("todos not found")
}

func (f *FakeQuerier) GetUserByEmail(ctx context.Context, username string) (db.User, error) {
	for id, user := range f.Users {
		if user.Email == username {
			return db.User{
				ID:           id,
				Email:        user.Email,
				PasswordHash: user.PasswordHash,
			}, nil
		}
	}
	return db.User{}, errors.New("user not found")
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
	return db.Session(db.Session{
		ID:        session.ID,
		UserID:    session.UserID,
		ExpiresAt: session.ExpiresAt,
		CreatedAt: sql.NullString{String: time.Now().Format(time.RFC3339), Valid: true},
	}), nil
}

func (f *FakeQuerier) UpdateTodo(ctx context.Context, arg db.UpdateTodoParams) (db.Todo, error) {
	if arg.ID == 0 || arg.UserID == 0 || arg.Name == "" {
		return db.Todo{}, errors.New("invalid todo parameters")
	}
	todo, exists := f.Todos[arg.ID]
	if !exists {
		return db.Todo{}, errors.New("todo not found")
	}
	todo.Name = arg.Name
	todo.Description = arg.Description
	todo.UserID = arg.UserID

	f.Todos[arg.ID] = todo
	return db.Todo{
		Name:        todo.Name,
		Description: todo.Description,
		UserID:      todo.UserID,
	}, nil
}

func (f *FakeQuerier) ValidateSessionToken(ctx context.Context, id string) (db.ValidateSessionTokenRow, error) {
	session, exists := f.Sessions[id]
	if !exists {
		return db.ValidateSessionTokenRow{}, errors.New("session not found")
	}
	return db.ValidateSessionTokenRow{
		ID:            session.ID,
		UserID:        session.UserID,
		ExpiresAt:     session.ExpiresAt,
		EmailVerified: 0,
	}, nil
}

func (f *FakeQuerier) InsertUserEmailVerificationRequest(ctx context.Context, arg db.InsertUserEmailVerificationRequestParams) (db.EmailVerificationRequest, error) {
	if arg.UserID == 0 || arg.Code == "" {
		return db.EmailVerificationRequest{}, errors.New("invalid email verification request parameters")
	}
	emailVerificationRequest := db.EmailVerificationRequest(arg)
	f.EmailVerificationRequests[arg.UserID] = arg
	return emailVerificationRequest, nil
}

func (f *FakeQuerier) DeleteUserEmailVerificationRequest(ctx context.Context, userId int64) error {
	if _, exists := f.EmailVerificationRequests[userId]; !exists {
		return errors.New("email verification request not found")
	}
	delete(f.EmailVerificationRequests, userId)
	return nil
}

func (f *FakeQuerier) GetUserEmailVerificationRequest(ctx context.Context, userId int64) (db.EmailVerificationRequest, error) {
	emailVerificationRequest, exists := f.EmailVerificationRequests[userId]
	if !exists {
		return db.EmailVerificationRequest{}, errors.New("email verification request not found")
	}
	return db.EmailVerificationRequest(emailVerificationRequest), nil
}

func (f *FakeQuerier) SetUserEmailVerified(ctx context.Context, userId int64) error {
	if _, exists := f.Users[userId]; !exists {
		return errors.New("user not found")
	}
	return nil
}

func (f *FakeQuerier) ValidateEmailVerificationRequest(ctx context.Context, arg db.ValidateEmailVerificationRequestParams) (db.EmailVerificationRequest, error) {
	if arg.UserID == 0 || arg.Code == "" {
		return db.EmailVerificationRequest{}, errors.New("invalid email verification request parameters")
	}
	emailVerificationRequest, exists := f.EmailVerificationRequests[arg.UserID]
	if !exists {
		return db.EmailVerificationRequest{}, errors.New("email verification request not found")
	}
	if emailVerificationRequest.Code != arg.Code {
		return db.EmailVerificationRequest{}, errors.New("invalid code")
	}
	if emailVerificationRequest.ExpiresAt < time.Now().Unix() {
		return db.EmailVerificationRequest{}, errors.New("email verification request expired")
	}
	return db.EmailVerificationRequest(emailVerificationRequest), nil
}
