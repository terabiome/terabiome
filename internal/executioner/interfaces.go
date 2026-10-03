package executioner

import (
	"context"

	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
)

type CommandExecutor interface {
	Execute(ctx context.Context, input *shellexecutors.CommandInput) shellexecutors.CommandOutput
	Name() string
}
