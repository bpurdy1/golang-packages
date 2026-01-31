package sloglogger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
)

// Config represents the settings populated by caarlos0/env
type Config struct {
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
	JSON      bool   `env:"LOG_JSON" envDefault:"false"`
	AddSource bool   `env:"LOG_SOURCE" envDefault:"false"`
}

type Option func(*Config)

func WithLevel(level string) Option {
	return func(c *Config) {
		c.LogLevel = level
	}
}

func WithJSON(json bool) Option {
	return func(c *Config) {
		c.JSON = json
	}
}

func WithSource(addSource bool) Option {
	return func(c *Config) {
		c.AddSource = addSource
	}
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SetGlobal(opts ...Option) error {
	cfg, err := NewConfig()
	if err != nil {
		return err
	}
	for _, opt := range opts {
		opt(cfg)
	}

	logger := NewLogger(cfg)
	slog.SetDefault(logger)

	return nil
}

func NewLogger(cfg *Config) *slog.Logger {
	if cfg == nil {
		return slog.Default()
	}

	level := parseLevel(cfg.LogLevel)

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if cfg.JSON {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type loggerKey struct{} // context internal key

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// WithContext wraps the logger into the context and returns the new context.
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// Log level wrappers

func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func Debugf(format string, fmtArgs ...any) {
	slog.Debug(fmt.Sprintf(format, fmtArgs...))
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Infof(format string, fmtArgs ...any) {
	slog.Info(fmt.Sprintf(format, fmtArgs...))
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func Warnf(format string, fmtArgs ...any) {
	slog.Warn(fmt.Sprintf(format, fmtArgs...))
}

func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

func Errorf(format string, fmtArgs ...any) {
	slog.Error(fmt.Sprintf(format, fmtArgs...))
}

// ErrorWithErr logs an error with the error attached
func ErrorWithErr(err error, msg string, args ...any) {
	args = append(args, "error", err)
	slog.Error(msg, args...)
}

// ErrorWithErrf logs a formatted error with the error attached
func ErrorWithErrf(err error, format string, fmtArgs ...any) {
	slog.Error(fmt.Sprintf(format, fmtArgs...), "error", err)
}
