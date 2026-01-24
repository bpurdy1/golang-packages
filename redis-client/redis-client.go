//go:generate mockgen -source=redis-client.go -destination=./mock/mock.go -package=mock
package redisclient

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/redis/go-redis/v9"
)

// Config holds the connection parameters
type Config struct {
	Addr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Password string `env:"REDIS_PASS"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

// NewConfig parses environment variables into the Config struct
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse redis config: %w", err)
	}
	return cfg, nil
}

// Client defines the contract for our Redis operations
type Client interface {
	redis.Cmdable
	Close() error
}

// NewClient initializes a new Redis client using the provided config
func NewClient(cfg *Config) Client {
	opt := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	return RedisClient{
		redis.NewClient(
			opt,
		),
	}
}

type Option func(*redis.Options)

func NewClientOptions(opts ...Option) Client {
	var rdcfg redis.Options
	for _, opt := range opts {
		opt(&rdcfg)
	}

	return RedisClient{
		redis.NewClient(&rdcfg),
	}

}

type RedisClient struct {
	*redis.Client
}
