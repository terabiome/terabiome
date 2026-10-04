package operations

import (
	"context"
)

// Operation is the atomic unit of work.
type Operation interface {
	Name() string
	PreCheck(env *OperationEnvironment) error
	Run(ctx context.Context, env *OperationEnvironment) error
}
