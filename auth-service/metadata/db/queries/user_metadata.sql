-- name: CreateUserMetadata :one
INSERT INTO user_metadata (user_id, key, value)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetUserMetadata :one
SELECT * FROM user_metadata
WHERE user_id = ? AND key = ? LIMIT 1;

-- name: GetUserMetadataByID :one
SELECT * FROM user_metadata
WHERE id = ? LIMIT 1;

-- name: ListUserMetadata :many
SELECT * FROM user_metadata
WHERE user_id = ?
ORDER BY key ASC;

-- name: ListAllUserMetadata :many
SELECT * FROM user_metadata
ORDER BY user_id, key ASC
LIMIT ? OFFSET ?;

-- name: UpdateUserMetadata :one
UPDATE user_metadata
SET value = ?
WHERE user_id = ? AND key = ?
RETURNING *;

-- name: UpsertUserMetadata :one
INSERT INTO user_metadata (user_id, key, value)
VALUES (?, ?, ?)
ON CONFLICT(user_id, key) DO UPDATE SET value = excluded.value
RETURNING *;

-- name: DeleteUserMetadata :exec
DELETE FROM user_metadata
WHERE user_id = ? AND key = ?;

-- name: DeleteAllUserMetadata :exec
DELETE FROM user_metadata
WHERE user_id = ?;

-- name: CountUserMetadata :one
SELECT COUNT(*) FROM user_metadata
WHERE user_id = ?;
