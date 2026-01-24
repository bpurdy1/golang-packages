package account

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
)

type UserService struct {
	*Queries
}

func NewUserService(db DBTX) *UserService {
	return &UserService{
		New(db),
	}
}

type CreateUserInput struct {
	FirstName string
	LastName  string
	Username  string
	Password  string
	Email     string
}

type UpdateUserInput struct {
	FirstName string
	LastName  string
	Username  string
	Email     string
}

func (s *UserService) CreateUser(ctx context.Context, input CreateUserInput) (User, error) {
	if input.Username == "" || input.Password == "" || input.Email == "" {
		return User{}, ErrInvalidInput
	}

	// Check if username already exists
	exists, err := s.Queries.UserExistsByUsername(ctx, input.Username)
	if err != nil {
		return User{}, fmt.Errorf("failed to check username existence: %w", err)
	}
	if exists == 1 {
		return User{}, fmt.Errorf("%w: username already taken", ErrUserAlreadyExists)
	}

	// Check if email already exists
	exists, err = s.Queries.UserExistsByEmail(ctx, input.Email)
	if err != nil {
		return User{}, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists == 1 {
		return User{}, fmt.Errorf("%w: email already registered", ErrUserAlreadyExists)
	}

	// Generate password hash
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate username+password hash
	usernamePasswordHash := generateUsernamePasswordHash(input.Username, input.Password)

	// Create user
	user, err := s.Queries.CreateUser(ctx, CreateUserParams{
		Uuid:                 uuid.New().String(),
		FirstName:            input.FirstName,
		LastName:             input.LastName,
		Username:             input.Username,
		PasswordHash:         passwordHash,
		UsernamePasswordHash: usernamePasswordHash,
		Email:                input.Email,
	})
	if err != nil {
		return User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id int64) (User, error) {
	user, err := s.Queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *UserService) GetUserByUUID(ctx context.Context, userUUID string) (User, error) {
	user, err := s.Queries.GetUserByUUID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (User, error) {
	user, err := s.Queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (User, error) {
	user, err := s.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int64) ([]User, error) {
	users, err := s.Queries.ListUsers(ctx, ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id int64, input UpdateUserInput) (User, error) {
	// Check if user exists
	_, err := s.Queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}

	user, err := s.Queries.UpdateUser(ctx, UpdateUserParams{
		ID:        id,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Username:  input.Username,
		Email:     input.Email,
	})
	if err != nil {
		return User{}, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (s *UserService) UpdatePassword(ctx context.Context, id int64, newPassword string) (User, error) {
	// Get existing user
	existingUser, err := s.Queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate new password hash
	passwordHash, err := hashPassword(newPassword)
	if err != nil {
		return User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate new username+password hash
	usernamePasswordHash := generateUsernamePasswordHash(existingUser.Username, newPassword)

	user, err := s.Queries.UpdateUserPassword(ctx, UpdateUserPasswordParams{
		ID:                   id,
		PasswordHash:         passwordHash,
		UsernamePasswordHash: usernamePasswordHash,
	})
	if err != nil {
		return User{}, fmt.Errorf("failed to update password: %w", err)
	}

	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	err := s.Queries.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (s *UserService) DeleteUserByUUID(ctx context.Context, userUUID string) error {
	err := s.Queries.DeleteUserByUUID(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (s *UserService) Authenticate(ctx context.Context, username, password string) (User, error) {
	user, err := s.Queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}

	if !checkPassword(password, user.PasswordHash) {
		return User{}, ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserService) CountUsers(ctx context.Context) (int64, error) {
	count, err := s.Queries.CountUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// hashPassword creates a bcrypt hash of the password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// checkPassword compares a password with a bcrypt hash
func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateUsernamePasswordHash creates a SHA256 hash of username+password
func generateUsernamePasswordHash(username, password string) string {
	combined := username + ":" + password
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}
