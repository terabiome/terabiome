package operations

import (
	"context"
)

// Operation is the atomic unit of work.
type Operation interface {
	Name() string
	SetParameters(parameters map[string]any) error
	PreCheck(env *OperationEnvironment) error
	Run(ctx context.Context, env *OperationEnvironment) error
}
