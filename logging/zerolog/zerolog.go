package zerologlogger

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// Get the last two parts of the path (e.g., cmd/main.go)
	ShortCallerMarshalFunc = func(_ uintptr, file string, line int) string {
		// Or use filepath.Base(file) for just "main.go"
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				// Change '2' to '3' or '4' depending on how many parent
				// directories you want to see.
				short = file[i+1:]
				break
			}
		}
		return short + ":" + strconv.Itoa(line)
	}
	FileBaseCallerMarshalFunc = func(_ uintptr, file string, line int) string {
		return file + ":" + strconv.Itoa(line)
	}
)

// Config represents the settings populated by caarlos0/env
type Config struct {
	LogLevel          string `env:"LOG_LEVEL" envDefault:"info"`
	ConsoleWriter     bool   `env:"LOG_CONSOLE" envDefault:"false"`
	CallerMarshalFunc func(pc uintptr, file string, line int) string
}
type Option func(*Config)

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

	newLogger := NewLogger(cfg)
	log.Logger = newLogger
	zerolog.DefaultContextLogger = &newLogger

	return nil
}

func NewLogger(cfg *Config) zerolog.Logger {
	if cfg == nil {
		return log.Logger
	}
	if cfg.CallerMarshalFunc != nil {
		zerolog.CallerMarshalFunc = cfg.CallerMarshalFunc
	} else {
		// Default ShortCallerMarshalFunc
		zerolog.CallerMarshalFunc = ShortCallerMarshalFunc
	}

	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	var out zerolog.LevelWriter
	if cfg.ConsoleWriter {
		out = zerolog.MultiLevelWriter(
			zerolog.ConsoleWriter{
				Out:        os.Stderr,
				TimeFormat: time.RFC3339,
			})
	} else {
		out = zerolog.MultiLevelWriter(os.Stderr)
	}

	newlogger := zerolog.
		New(out).
		Level(level).
		With().
		Timestamp().
		Caller(). // Adds file and line number
		Logger()
	return newlogger
}

func FromContext(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

func WithContext(ctx context.Context, logger zerolog.Logger) context.Context {
	return logger.WithContext(ctx)
}

// Set Fields in Context
func WithFields(ctx context.Context, fields map[string]any) context.Context {
	l := FromContext(ctx).With()
	for k, v := range fields {
		l = l.Interface(k, v)
	}
	newLogger := l.Logger()
	return newLogger.WithContext(ctx)
}

func Debug(msg string) {
	log.Debug().Msg(msg)
}

func Debugf(format string, args ...any) {
	log.Debug().Msgf(format, args...)
}

func Info(msg string) {
	log.Info().Msg(msg)
}

func Infof(format string, args ...any) {
	log.Info().Msgf(format, args...)
}

func Warn(msg string) {
	log.Warn().Msg(msg)
}

func Warnf(format string, args ...any) {
	log.Warn().Msgf(format, args...)
}

func Error(err error, msg string) {
	log.Error().Err(err).Msg(msg)
}

func Errorf(err error, format string, args ...any) {
	log.Error().Err(err).Msgf(format, args...)
}

func Fatal(msg string) {
	log.Fatal().Msg(msg)
}

func Fatalf(format string, args ...any) {
	log.Fatal().Msgf(format, args...)
}
