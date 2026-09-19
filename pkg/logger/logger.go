package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/terabiome/infrastructures/internal/config"
)

type Logger struct {
	cfg config.LoggerConfig
	*slog.Logger
}

func New(cfg config.LoggerConfig) *Logger {
	var logLevel slog.Level

	switch strings.ToLower(cfg.Level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler

	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return &Logger{
		cfg:    cfg,
		Logger: slog.New(handler),
	}
}

func NewWithComponent(cfg config.LoggerConfig) *Logger {
	logger := New(cfg)
	logger.Logger = logger.Logger.With(
		slog.String("component", cfg.Component),
	)
	return logger
}
