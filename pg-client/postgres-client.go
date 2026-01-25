//go:generate mockgen -source=postgres-client.go -destination=./mock/mock.go -package=mock
package pgclient

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/caarlos0/env/v11"
	_ "github.com/lib/pq" // Postgres driver
)

//go:embed queries/*.sql
var queryFS embed.FS

type QueryLib struct {
	Query1 string
	Query2 string
	Query3 string
}

var lib = QueryLib{
	Query1: read("queries/file1.sql"),
	Query2: read("queries/file2.sql"),
	Query3: read("queries/file3.sql"),
}

func read(file string) string {
	b, err := queryFS.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(b)
}

type Config struct {
	Host    string `env:"DB_HOST" envDefault:"localhost"`
	Port    int    `env:"DB_PORT" envDefault:"5432"`
	User    string `env:"DB_USER" envDefault:"postgres"`
	Pass    string `env:"DB_PASS"`
	Name    string `env:"DB_NAME"`
	SSLMode string `env:"DB_SSLMODE" envDefault:"disable"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse pg config: %w", err)
	}
	return cfg, nil
}

type Client interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Ping() error
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Close() error
}

type PostgresClient struct {
	*sql.DB
}

func NewClient(cfg *Config) (Client, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.Name, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresClient{db}, nil
}

// ExampleQuery1 demonstrates using an embedded SQL query from QueryLib.
func (c *PostgresClient) ExampleQuery1(ctx context.Context, args ...any) (*sql.Rows, error) {
	return c.QueryContext(ctx, lib.Query1, args...)
}

type ExampleQuery2Request struct {
	ID   int
	Name string
}

type ExampleQuery2Response struct {
	ID        int
	Name      string
	CreatedAt string
}

// ExampleQuery2 demonstrates using request/response structs with an embedded SQL query.
func (c *PostgresClient) ExampleQuery2(ctx context.Context, req ExampleQuery2Request) (*ExampleQuery2Response, error) {
	row := c.QueryRowContext(ctx, lib.Query2, req.ID, req.Name)

	var resp ExampleQuery2Response
	if err := row.Scan(&resp.ID, &resp.Name, &resp.CreatedAt); err != nil {
		return nil, err
	}
	return &resp, nil
}
