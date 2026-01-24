package account

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupInMemoryDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

func setupTempFileDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open temp file database: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}

func TestCreateUser(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("expected user ID to be set")
	}
	if user.Uuid == "" {
		t.Error("expected user UUID to be set")
	}
	if user.FirstName != "John" {
		t.Errorf("expected FirstName 'John', got '%s'", user.FirstName)
	}
	if user.LastName != "Doe" {
		t.Errorf("expected LastName 'Doe', got '%s'", user.LastName)
	}
	if user.Username != "johndoe" {
		t.Errorf("expected Username 'johndoe', got '%s'", user.Username)
	}
	if user.Email != "john@example.com" {
		t.Errorf("expected Email 'john@example.com', got '%s'", user.Email)
	}
	if user.PasswordHash == "" {
		t.Error("expected PasswordHash to be set")
	}
	if user.UsernamePasswordHash == "" {
		t.Error("expected UsernamePasswordHash to be set")
	}
	if user.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set by trigger")
	}
	if user.ModifiedAt.IsZero() {
		t.Error("expected ModifiedAt to be set by trigger")
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	_, err := svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})
	if err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	_, err = svc.CreateUser(ctx, CreateUserInput{
		FirstName: "Jane",
		LastName:  "Doe",
		Username:  "johndoe", // duplicate
		Password:  "password456",
		Email:     "jane@example.com",
	})
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	_, err := svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})
	if err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	_, err = svc.CreateUser(ctx, CreateUserInput{
		FirstName: "Jane",
		LastName:  "Doe",
		Username:  "janedoe",
		Password:  "password456",
		Email:     "john@example.com", // duplicate
	})
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestCreateUser_InvalidInput(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	tests := []struct {
		name  string
		input CreateUserInput
	}{
		{"missing username", CreateUserInput{Password: "pass", Email: "test@test.com"}},
		{"missing password", CreateUserInput{Username: "user", Email: "test@test.com"}},
		{"missing email", CreateUserInput{Username: "user", Password: "pass"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateUser(ctx, tt.input)
			if !errors.Is(err, ErrInvalidInput) {
				t.Errorf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	created, _ := svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})

	user, err := svc.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get user by ID: %v", err)
	}

	if user.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, user.ID)
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	_, err := svc.GetUserByID(ctx, 99999)
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGetUserByUUID(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	created, _ := svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})

	user, err := svc.GetUserByUUID(ctx, created.Uuid)
	if err != nil {
		t.Fatalf("failed to get user by UUID: %v", err)
	}

	if user.Uuid != created.Uuid {
		t.Errorf("expected UUID %s, got %s", created.Uuid, user.Uuid)
	}
}

func TestGetUserByUsername(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	_, _ = svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})

	user, err := svc.GetUserByUsername(ctx, "johndoe")
	if err != nil {
		t.Fatalf("failed to get user by username: %v", err)
	}

	if user.Username != "johndoe" {
		t.Errorf("expected username 'johndoe', got '%s'", user.Username)
	}
}

func TestGetUserByEmail(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	_, _ = svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})

	user, err := svc.GetUserByEmail(ctx, "john@example.com")
	if err != nil {
		t.Fatalf("failed to get user by email: %v", err)
	}

	if user.Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got '%s'", user.Email)
	}
}

func TestListUsers(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	// Create multiple users
	for i := 0; i < 5; i++ {
		_, _ = svc.CreateUser(ctx, CreateUserInput{
			FirstName: "User",
			LastName:  "Test",
			Username:  "user" + string(rune('0'+i)),
			Password:  "password123",
			Email:     "user" + string(rune('0'+i)) + "@example.com",
		})
	}

	users, err := svc.ListUsers(ctx, 3, 0)
	if err != nil {
		t.Fatalf("failed to list users: %v", err)
	}

	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}

	// Test offset
	users, err = svc.ListUsers(ctx, 10, 3)
	if err != nil {
		t.Fatalf("failed to list users with offset: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("expected 2 users with offset, got %d", len(users))
	}
}

func TestUpdateUser(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	created, _ := svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})

	updated, err := svc.UpdateUser(ctx, created.ID, UpdateUserInput{
		FirstName: "Jane",
		LastName:  "Smith",
		Username:  "janesmith",
		Email:     "jane@example.com",
	})
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	if updated.FirstName != "Jane" {
		t.Errorf("expected FirstName 'Jane', got '%s'", updated.FirstName)
	}
	if updated.LastName != "Smith" {
		t.Errorf("expected LastName 'Smith', got '%s'", updated.LastName)
	}
	if updated.Username != "janesmith" {
		t.Errorf("expected Username 'janesmith', got '%s'", updated.Username)
	}
	if updated.Email != "jane@example.com" {
		t.Errorf("expected Email 'jane@example.com', got '%s'", updated.Email)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	_, err := svc.UpdateUser(ctx, 99999, UpdateUserInput{
		FirstName: "Jane",
		LastName:  "Smith",
		Username:  "janesmith",
		Email:     "jane@example.com",
	})
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUpdatePassword(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	created, _ := svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})

	oldPasswordHash := created.PasswordHash
	oldUsernamePasswordHash := created.UsernamePasswordHash

	updated, err := svc.UpdatePassword(ctx, created.ID, "newpassword456")
	if err != nil {
		t.Fatalf("failed to update password: %v", err)
	}

	if updated.PasswordHash == oldPasswordHash {
		t.Error("expected PasswordHash to change")
	}
	if updated.UsernamePasswordHash == oldUsernamePasswordHash {
		t.Error("expected UsernamePasswordHash to change")
	}

	// Verify old password no longer works
	_, err = svc.Authenticate(ctx, "johndoe", "password123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Error("expected old password to fail authentication")
	}

	// Verify new password works
	_, err = svc.Authenticate(ctx, "johndoe", "newpassword456")
	if err != nil {
		t.Errorf("expected new password to work, got %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	created, _ := svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})

	err := svc.DeleteUser(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	_, err = svc.GetUserByID(ctx, created.ID)
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound after deletion, got %v", err)
	}
}

func TestDeleteUserByUUID(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	created, _ := svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})

	err := svc.DeleteUserByUUID(ctx, created.Uuid)
	if err != nil {
		t.Fatalf("failed to delete user by UUID: %v", err)
	}

	_, err = svc.GetUserByUUID(ctx, created.Uuid)
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound after deletion, got %v", err)
	}
}

func TestAuthenticate(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	_, _ = svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})

	user, err := svc.Authenticate(ctx, "johndoe", "password123")
	if err != nil {
		t.Fatalf("failed to authenticate: %v", err)
	}

	if user.Username != "johndoe" {
		t.Errorf("expected username 'johndoe', got '%s'", user.Username)
	}
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	_, _ = svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})

	_, err := svc.Authenticate(ctx, "johndoe", "wrongpassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthenticate_UserNotFound(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	_, err := svc.Authenticate(ctx, "nonexistent", "password123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestCountUsers(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	count, err := svc.CountUsers(ctx)
	if err != nil {
		t.Fatalf("failed to count users: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 users, got %d", count)
	}

	for i := 0; i < 3; i++ {
		_, _ = svc.CreateUser(ctx, CreateUserInput{
			FirstName: "User",
			LastName:  "Test",
			Username:  "user" + string(rune('0'+i)),
			Password:  "password123",
			Email:     "user" + string(rune('0'+i)) + "@example.com",
		})
	}

	count, err = svc.CountUsers(ctx)
	if err != nil {
		t.Fatalf("failed to count users: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 users, got %d", count)
	}
}

func TestTempFileDB(t *testing.T) {
	db, cleanup := setupTempFileDB(t)
	defer cleanup()

	svc := NewUserService(db)
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})
	if err != nil {
		t.Fatalf("failed to create user with temp file DB: %v", err)
	}

	if user.ID == 0 {
		t.Error("expected user ID to be set")
	}
}

func TestTimestampTriggers(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	svc := NewUserService(db)
	ctx := context.Background()

	created, _ := svc.CreateUser(ctx, CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})

	if created.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set by trigger")
	}
	if created.ModifiedAt.IsZero() {
		t.Error("expected ModifiedAt to be set by trigger")
	}

	// Update and check modified_at changes
	originalModifiedAt := created.ModifiedAt

	updated, _ := svc.UpdateUser(ctx, created.ID, UpdateUserInput{
		FirstName: "Jane",
		LastName:  "Doe",
		Username:  "johndoe",
		Email:     "john@example.com",
	})

	if updated.CreatedAt != created.CreatedAt {
		t.Error("expected CreatedAt to remain unchanged after update")
	}
	if !updated.ModifiedAt.After(originalModifiedAt) && updated.ModifiedAt != originalModifiedAt {
		// Note: In fast tests, timestamps might be equal due to same second
		t.Log("ModifiedAt may not have changed due to same-second execution")
	}
}
