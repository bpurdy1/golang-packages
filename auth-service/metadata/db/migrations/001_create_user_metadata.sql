-- +goose Up
-- Create user_metadata table for storing arbitrary key-value pairs per user
CREATE TABLE IF NOT EXISTS user_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, key)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_user_metadata_user_id ON user_metadata(user_id);
CREATE INDEX IF NOT EXISTS idx_user_metadata_key ON user_metadata(key);

-- +goose StatementBegin
-- Trigger to set timestamps on insert
CREATE TRIGGER IF NOT EXISTS set_user_metadata_created_at
AFTER INSERT ON user_metadata
BEGIN
    UPDATE user_metadata
    SET created_at = CURRENT_TIMESTAMP,
        modified_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
-- Trigger to update modified_at on update
CREATE TRIGGER IF NOT EXISTS set_user_metadata_modified_at
AFTER UPDATE ON user_metadata
BEGIN
    UPDATE user_metadata
    SET modified_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS set_user_metadata_modified_at;
DROP TRIGGER IF EXISTS set_user_metadata_created_at;
DROP INDEX IF EXISTS idx_user_metadata_key;
DROP INDEX IF EXISTS idx_user_metadata_user_id;
DROP TABLE IF EXISTS user_metadata;
