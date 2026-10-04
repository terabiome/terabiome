package executor

import (
	"context"
	"fmt"
	"sync"

	yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"
	"github.com/terabiome/terabiome/internal/operations"
)

// WorkflowExecutor handles the top-level coordination of a UnitOfWork.
type WorkflowExecutor struct {
	env *operations.OperationEnvironment
}

func NewWorkflowExecutor(env *operations.OperationEnvironment) *WorkflowExecutor {
	return &WorkflowExecutor{env: env}
}

// Execute takes pre-sorted batches and runs them.
// Batches are sequential; sequences within a batch are concurrent.
func (w *WorkflowExecutor) Execute(ctx context.Context, batches [][]yamlcontracts.UnitOfWorkSequence) error {
	for i, batch := range batches {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := w.runBatch(ctx, batch); err != nil {
			return fmt.Errorf("batch %d failed: %w", i, err)
		}
	}
	return nil
}

func (w *WorkflowExecutor) runBatch(ctx context.Context, sequences []yamlcontracts.UnitOfWorkSequence) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(sequences))

	for _, seq := range sequences {
		wg.Add(1)
		go func(s yamlcontracts.UnitOfWorkSequence) {
			defer wg.Done()
			// Reuse our SequenceExecutor intrinsic
			exec := NewSequenceExecutor(w.env)
			if err := exec.Execute(ctx, s); err != nil {
				errChan <- err
			}
		}(seq)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}
