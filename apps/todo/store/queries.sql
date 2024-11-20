-- name: GetTodo :one
SELECT * FROM todos WHERE id = ? AND user_id = ?;

-- name: GetTodos :many
SELECT * FROM todos WHERE user_id = ?;

-- name: CreateTodo :one
INSERT INTO todos (name, user_id, description) VALUES (?, ?, ?) RETURNING *;

-- name: UpdateTodo :one
UPDATE todos SET name = ?, description = ? WHERE id = ? AND user_id = ? RETURNING *;

-- name: DeleteTodo :exec
DELETE FROM todos WHERE id = ? AND user_id = ?;

-- name: CreateSession :one
INSERT INTO session (id, user_id, expires_at) VALUES (?, ?, ?) RETURNING *;

-- name: ValidateSessionToken :one
SELECT s.id, s.user_id as user_id, s.expires_at, u.email, u.email_verified
FROM session s 
INNER JOIN user u ON u.id = s.user_id 
WHERE s.id = ?;

-- name: DeleteSession :exec
DELETE FROM session WHERE id = ?;

-- name: UpdateSession :one
UPDATE session SET expires_at = ? WHERE id = ? RETURNING *;

-- name: CreateUser :one
INSERT INTO user (email, password_hash) VALUES (?, ?) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM user WHERE email = ?;

-- name: InsertUserEmailVerificationRequest :one
INSERT INTO email_verification_request (user_id, created_at, expires_at, code) 
VALUES (?, ?, ?, ?)
ON CONFLICT(user_id) DO UPDATE SET created_at = EXCLUDED.created_at, expires_at = EXCLUDED.expires_at, code = EXCLUDED.code
RETURNING *;

-- name: GetUserEmailVerificationRequest :one
SELECT * FROM email_verification_request WHERE user_id = ?;

-- name: DeleteUserEmailVerificationRequest :exec
DELETE FROM email_verification_request WHERE user_id = ?; 

-- name: ValidateEmailVerificationRequest :one
DELETE FROM email_verification_request WHERE user_id = ? AND code = ? AND expires_at > ? RETURNING *;

-- name: SetUserEmailVerified :exec
UPDATE user SET email_verified = 1 WHERE id = ?;