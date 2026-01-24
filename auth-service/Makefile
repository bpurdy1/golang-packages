.PHONY: help migrate-up migrate-down migrate-status migrate-create migrate-reset sqlc test run

# Database configuration
DB_PATH ?= ./users.db
MIGRATIONS_DIR = ./account/db/migrations

help:
	@echo "Usage:"
	@echo "  make migrate-up        - Run all pending migrations"
	@echo "  make migrate-down      - Rollback the last migration"
	@echo "  make migrate-status    - Show migration status"
	@echo "  make migrate-create    - Create a new migration (use NAME=migration_name)"
	@echo "  make migrate-reset     - Rollback all migrations and re-run them"
	@echo "  make sqlc              - Generate sqlc code"
	@echo "  make test              - Run tests"
	@echo "  make run               - Run the example"
	@echo ""
	@echo "Examples:"
	@echo "  make migrate-create NAME=add_user_roles"
	@echo "  make migrate-up DB_PATH=./prod.db"

# Run all pending migrations
migrate-up:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) up

# Rollback the last migration
migrate-down:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) down

# Show migration status
migrate-status:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) status

# Create a new migration file
migrate-create:
ifndef NAME
	$(error NAME is required. Usage: make migrate-create NAME=migration_name)
endif
	goose -dir $(MIGRATIONS_DIR) create $(NAME) sql

# Rollback all and re-run migrations
migrate-reset:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) reset
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) up

# Generate sqlc code
sqlc:
	sqlc generate

# Run tests
test:
	go test -v ./account/...

# Run the example
run:
	go run ./cmd/example
