package executioner

import (
	"context"

	"github.com/terabiome/infrastructures/internal/executor"
)

type CommandExecutor interface {
	Execute(ctx context.Context, input *executor.CommandInput) executor.CommandOutput
	Name() string
}

type Logger interface {
	Debug(msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
	Error(msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	Info(msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	Warn(msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
}
