package account

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed db/migrations/*.sql
var migrations embed.FS

// TableName is the goose version table name for this package
const TableName = "goose_db_version_account"

// Migrate runs all pending database migrations using goose
func Migrate(db *sql.DB) error {
	goose.SetBaseFS(migrations)
	goose.SetTableName(TableName)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	return goose.Up(db, "db/migrations")
}

// MigrateDown rolls back the last migration
func MigrateDown(db *sql.DB) error {
	goose.SetBaseFS(migrations)
	goose.SetTableName(TableName)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	return goose.Down(db, "db/migrations")
}

// MigrateReset rolls back all migrations and re-runs them
func MigrateReset(db *sql.DB) error {
	goose.SetBaseFS(migrations)
	goose.SetTableName(TableName)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	if err := goose.Reset(db, "db/migrations"); err != nil {
		return err
	}

	return goose.Up(db, "db/migrations")
}
