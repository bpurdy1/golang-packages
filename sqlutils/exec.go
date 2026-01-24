package sqlutils

import (
	"context"
	"database/sql"
)

// ExecGetID executes an insert and returns the last inserted ID
func ExecGetID(ctx context.Context, db *sql.DB, query string, args ...any) (int64, error) {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// ExecGetIDTx executes an insert within a transaction and returns the last inserted ID
func ExecGetIDTx(ctx context.Context, tx *sql.Tx, query string, args ...any) (int64, error) {
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
