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

-- name: CreateSession :one
INSERT INTO session (id, user_id, expires_at) VALUES (?, ?, ?) RETURNING *;

-- name: ValidateSessionToken :one
SELECT s.id, s.user_id as session_user_id, s.expires_at, u.id AS user_id, u.username 
FROM session s 
INNER JOIN user u ON u.id = s.user_id 
WHERE s.id = ?;

-- name: DeleteSession :exec
DELETE FROM session WHERE id = ?;

-- name: UpdateSession :one
UPDATE session SET expires_at = ? WHERE id = ? RETURNING *;

-- name: CreateUser :one
INSERT INTO user (username, password_hash) VALUES (?, ?) RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM user WHERE username = ?;



