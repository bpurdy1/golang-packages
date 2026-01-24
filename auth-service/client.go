package authservice

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/bpurdy1/auth-service/account"
	"github.com/bpurdy1/auth-service/config"
	"github.com/bpurdy1/auth-service/metadata"
)

// Config wraps the database configuration
type Config = config.Config

// Option is a function that modifies the config
type Option func(*Config)

// Client provides unified access to all services
type Client struct {
	db       *sql.DB
	cfg      *Config
	Users    *account.UserService
	Metadata metadata.MetadataService
}

// NewClient creates a new client with optional configuration
func NewClient(opts ...Option) (*Client, error) {
	cfg := &Config{
		DBPath:          ":memory:",
		DBMaxOpenConns:  25,
		DBMaxIdleConns:  5,
		DBJournalMode:   "WAL",
		DBBusyTimeoutMs: 5000,
		DBCacheSize:     -2000,
		DBSynchronous:   "NORMAL",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	db, err := cfg.OpenDB()
	if err != nil {
		return nil, err
	}

	// Run migrations
	if err := account.Migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	if err := metadata.Migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Client{
		db:       db,
		cfg:      cfg,
		Users:    account.NewUserService(db),
		Metadata: metadata.NewMetadataService(db),
	}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

// DB returns the underlying database connection
func (c *Client) DB() *sql.DB {
	return c.db
}

// UserWithMetadata represents a user with their metadata
type UserWithMetadata struct {
	account.User
	Metadata map[string]string `json:"metadata"`
}

// Re-export types for convenience
type (
	CreateUserInput  = account.CreateUserInput
	UpdateUserInput  = account.UpdateUserInput
	User             = account.User
	SetMetadataInput = metadata.SetMetadataInput
	UserMetadata     = metadata.UserMetadata
)

// GetUserWithMetadata fetches a user by ID with all their metadata
func (c *Client) GetUserWithMetadata(ctx context.Context, userID int64) (*UserWithMetadata, error) {
	user, err := c.Users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	meta, err := c.Metadata.AsMap(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserWithMetadata{
		User:     user,
		Metadata: meta,
	}, nil
}

// CreateUserWithMetadata creates a user and sets initial metadata
func (c *Client) CreateUserWithMetadata(ctx context.Context, input CreateUserInput, meta map[string]string) (*UserWithMetadata, error) {
	user, err := c.Users.CreateUser(ctx, input)
	if err != nil {
		return nil, err
	}

	for key, value := range meta {
		_, err := c.Metadata.Set(
			ctx,
			SetMetadataInput{
				UserID: user.ID,
				Key:    key,
				Value:  value,
			})
		if err != nil {
			return nil, err
		}
	}

	return &UserWithMetadata{
		User:     user,
		Metadata: meta,
	}, nil
}
