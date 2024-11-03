-- name: GetTodo :one
SELECT * FROM todos WHERE id = ?;

-- name: GetTodos :many
SELECT * FROM todos;

-- name: CreateTodo :one
INSERT INTO todos (name, description) VALUES (?, ?) RETURNING *;

-- name: UpdateTodo :one
UPDATE todos SET name = ?, description = ? WHERE id = ? RETURNING *;

-- name: DeleteTodo :exec
DELETE FROM todos WHERE id = ?;