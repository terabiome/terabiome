package operations

import (
	"context"

	"github.com/terabiome/terabiome/internal/writers"
)

type DemoOperation struct {
	parameters map[string]any
}

func NewDemoOperation() *DemoOperation {
	return &DemoOperation{}
}

// Name implements [operations.Operation].
func (ops *DemoOperation) Name() string {
	return string(OperationNameDemo)
}

// PreCheck implements [operations.Operation].
func (ops *DemoOperation) PreCheck(env *OperationEnvironment) error {
	return nil
}

// Run implements [operations.Operation].
func (ops *DemoOperation) Run(ctx context.Context, env *OperationEnvironment) error {
	env.Journaler.Log("terabiome!", writers.JournalerMetadata{})
	return nil
}

// SetParameters implements [operations.Operation].
func (ops *DemoOperation) SetParameters(parameters map[string]any) error {
	return nil
}
