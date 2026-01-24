package config

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	_ "github.com/mattn/go-sqlite3"
)

// Config holds the database configuration loaded from environment variables
type Config struct {
	// DBPath is the path to the SQLite database file
	// Use ":memory:" for in-memory database
	DBPath string `env:"DB_PATH" envDefault:"./data.db"`

	// DBMaxOpenConns sets the maximum number of open connections
	DBMaxOpenConns int `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`

	// DBMaxIdleConns sets the maximum number of idle connections
	DBMaxIdleConns int `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`

	// DBConnMaxLifetimeSecs sets the maximum lifetime of connections in seconds (0 = unlimited)
	DBConnMaxLifetimeSecs int `env:"DB_CONN_MAX_LIFETIME_SECS" envDefault:"0"`

	// DBJournalMode sets the SQLite journal mode (DELETE, TRUNCATE, PERSIST, MEMORY, WAL, OFF)
	DBJournalMode string `env:"DB_JOURNAL_MODE" envDefault:"WAL"`

	// DBBusyTimeoutMs sets the busy timeout in milliseconds
	DBBusyTimeoutMs int `env:"DB_BUSY_TIMEOUT_MS" envDefault:"5000"`

	// DBCacheSize sets the cache size in pages (negative = KB)
	DBCacheSize int `env:"DB_CACHE_SIZE" envDefault:"-2000"`

	// DBSynchronous sets the synchronous mode (OFF, NORMAL, FULL, EXTRA)
	DBSynchronous string `env:"DB_SYNCHRONOUS" envDefault:"NORMAL"`
}

// Load parses environment variables into a Config struct
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

// MustLoad loads the config and panics on error
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}

// DSN returns the data source name for the SQLite connection
func (c *Config) DSN() string {
	if c.DBPath == ":memory:" {
		return ":memory:?cache=shared"
	}
	return fmt.Sprintf("%s?_journal_mode=%s&_busy_timeout=%d&_cache_size=%d&_synchronous=%s",
		c.DBPath,
		c.DBJournalMode,
		c.DBBusyTimeoutMs,
		c.DBCacheSize,
		c.DBSynchronous,
	)
}

// OpenDB opens a database connection using the configuration
func (c *Config) OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", c.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(c.DBMaxOpenConns)
	db.SetMaxIdleConns(c.DBMaxIdleConns)

	if c.DBConnMaxLifetimeSecs > 0 {
		db.SetConnMaxLifetime(
			time.Duration(c.DBConnMaxLifetimeSecs) * time.Second,
		)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// MustOpenDB opens a database connection and panics on error
func (c *Config) MustOpenDB() *sql.DB {
	db, err := c.OpenDB()
	if err != nil {
		panic(err)
	}
	return db
}
