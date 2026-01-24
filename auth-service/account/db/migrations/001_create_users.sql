-- +goose Up
-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    username_password_hash TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for common lookups
CREATE INDEX IF NOT EXISTS idx_users_uuid ON users(uuid);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- +goose StatementBegin
-- Trigger to set created_at on insert (ensures consistent timestamp)
CREATE TRIGGER IF NOT EXISTS set_users_created_at
AFTER INSERT ON users
BEGIN
    UPDATE users
    SET created_at = CURRENT_TIMESTAMP,
        modified_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
-- Trigger to update modified_at on update
CREATE TRIGGER IF NOT EXISTS set_users_modified_at
AFTER UPDATE ON users
BEGIN
    UPDATE users
    SET modified_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS set_users_modified_at;
DROP TRIGGER IF EXISTS set_users_created_at;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_uuid;
DROP TABLE IF EXISTS users;
