package executioner

import (
	"context"

	"github.com/terabiome/infrastructures/internal/executor"
)

type Executor interface {
	Execute(ctx context.Context, input *executor.Input) executor.Output
	Name() string
}
