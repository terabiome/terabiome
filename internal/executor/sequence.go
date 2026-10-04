package executor

import (
	"context"
	"fmt"

	yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"
	"github.com/terabiome/terabiome/internal/operations"
	"github.com/terabiome/terabiome/internal/registry"
)

// SequenceExecutor is an intrinsic block that executes a single sequence of tasks.
type SequenceExecutor struct {
	env *operations.OperationEnvironment
}

func NewSequenceExecutor(env *operations.OperationEnvironment) *SequenceExecutor {
	return &SequenceExecutor{env: env}
}

// Execute runs all steps in the sequence sequentially.
func (e *SequenceExecutor) Execute(ctx context.Context, sequence yamlcontracts.UnitOfWorkSequence) error {
	for i, step := range sequence.Steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := e.runStep(ctx, step); err != nil {
			return fmt.Errorf("step %d (%s) failed: %w", i, step.Name, err)
		}
	}
	return nil
}

func (e *SequenceExecutor) runStep(ctx context.Context, step yamlcontracts.UnitOfWorkStep) error {
	// 1. Resolve Operation
	op, err := registry.Get(step.Name)
	if err != nil {
		return err
	}

	// 2. Inject Parameters
	if err := op.SetParameters(step.Parameters); err != nil {
		return fmt.Errorf("params invalid for %s: %w", step.Name, err)
	}

	// 3. PreCheck
	if err := op.PreCheck(e.env); err != nil {
		return fmt.Errorf("pre-check failed for %s: %w", step.Name, err)
	}

	// 4. Run
	return op.Run(ctx, e.env)
}
