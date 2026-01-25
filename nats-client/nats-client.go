//go:generate mockgen -source=nats-client.go -destination=./mock/mock.go -package=mock
package natsclient

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/nats-io/nats.go"
)

// Config holds the connection parameters for NATS
type Config struct {
	URL      string `env:"NATS_URL" envDefault:"nats://localhost:4222"`
	Token    string `env:"NATS_TOKEN"`
	User     string `env:"NATS_USER"`
	Password string `env:"NATS_PASS"`
}

// NewConfig parses environment variables into the Config struct
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse nats config: %w", err)
	}
	return cfg, nil
}

// Client defines the contract for NATS operations
type Client interface {
	Publish(subj string, data []byte) error
	Request(subj string, data []byte, timeout time.Duration) (*nats.Msg, error)
	Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error)
	QueueSubscribe(subj, queue string, cb nats.MsgHandler) (*nats.Subscription, error)
	Flush() error
	Close()
}

// NatsClient wraps the underlying nats.Conn to satisfy the Client interface
type NatsClient struct {
	*nats.Conn
}

// NewClient initializes a NATS client using the provided config
func NewClient(cfg *Config) (Client, error) {
	opts := nats.Options{
		Url:      cfg.URL,
		Token:    cfg.Token,
		User:     cfg.User,
		Password: cfg.Password,
	}

	nc, err := opts.Connect()
	if err != nil {
		return nil, err
	}

	return &NatsClient{nc}, nil
}

// Functional Options support
type Option func(*nats.Options)

func NewClientOptions(opts ...Option) (Client, error) {
	nopts := nats.GetDefaultOptions()
	for _, opt := range opts {
		opt(&nopts)
	}

	nc, err := nopts.Connect()
	if err != nil {
		return nil, err
	}

	return &NatsClient{nc}, nil
}
