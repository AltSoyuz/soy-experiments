// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package db

import (
	"context"
	"database/sql"
)

const createTodo = `-- name: CreateTodo :one
INSERT INTO todos (name, description) VALUES (?, ?) RETURNING id, name, description
`

type CreateTodoParams struct {
	Name        string
	Description sql.NullString
}

func (q *Queries) CreateTodo(ctx context.Context, arg CreateTodoParams) (Todo, error) {
	row := q.db.QueryRowContext(ctx, createTodo, arg.Name, arg.Description)
	var i Todo
	err := row.Scan(&i.ID, &i.Name, &i.Description)
	return i, err
}

const deleteTodo = `-- name: DeleteTodo :exec
DELETE FROM todos WHERE id = ?
`

func (q *Queries) DeleteTodo(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteTodo, id)
	return err
}

const getTodo = `-- name: GetTodo :one
SELECT id, name, description FROM todos WHERE id = ?
`

func (q *Queries) GetTodo(ctx context.Context, id int64) (Todo, error) {
	row := q.db.QueryRowContext(ctx, getTodo, id)
	var i Todo
	err := row.Scan(&i.ID, &i.Name, &i.Description)
	return i, err
}

const getTodos = `-- name: GetTodos :many
SELECT id, name, description FROM todos
`

func (q *Queries) GetTodos(ctx context.Context) ([]Todo, error) {
	rows, err := q.db.QueryContext(ctx, getTodos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Todo
	for rows.Next() {
		var i Todo
		if err := rows.Scan(&i.ID, &i.Name, &i.Description); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateTodo = `-- name: UpdateTodo :one
UPDATE todos SET name = ?, description = ? WHERE id = ? RETURNING id, name, description
`

type UpdateTodoParams struct {
	Name        string
	Description sql.NullString
	ID          int64
}

func (q *Queries) UpdateTodo(ctx context.Context, arg UpdateTodoParams) (Todo, error) {
	row := q.db.QueryRowContext(ctx, updateTodo, arg.Name, arg.Description, arg.ID)
	var i Todo
	err := row.Scan(&i.ID, &i.Name, &i.Description)
	return i, err
}