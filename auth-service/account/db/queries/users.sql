-- name: CreateUser :one
INSERT INTO users (
    uuid,
    first_name,
    last_name,
    username,
    password_hash,
    username_password_hash,
    email
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

-- name: GetUserByUUID :one
SELECT * FROM users
WHERE uuid = ? LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ? LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateUser :one
UPDATE users
SET first_name = ?,
    last_name = ?,
    username = ?,
    email = ?
WHERE id = ?
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users
SET password_hash = ?,
    username_password_hash = ?
WHERE id = ?
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;

-- name: DeleteUserByUUID :exec
DELETE FROM users
WHERE uuid = ?;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: UserExistsByUsername :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = ?) AS user_exists;

-- name: UserExistsByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = ?) AS user_exists;
