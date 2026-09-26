package executioner

import (
	"context"
	"errors"
	"fmt"

	"github.com/terabiome/infrastructures/internal/config"
	"github.com/terabiome/infrastructures/internal/models"
	"github.com/terabiome/infrastructures/pkg/logging"
)

type osExecutioner struct {
	// sub-executioners
	shellExecutioner   shellExecutioner
	networkExecutioner networkExecutioner
	diskExecutioner    diskExecutioner
	kernelExecutioner  kernelExecutioner
}

func NewOSExecutioner(executor CommandExecutor, cfg models.OSConfig) *osExecutioner {
	logger := logging.New(config.Get().Logger, map[string]any{
		"scope": "executioner",
		"name":  "os",
	})
	return &osExecutioner{
		shellExecutioner:   shellExecutioner{cfg.Shell, executor, logger},
		networkExecutioner: networkExecutioner{cfg.Network, executor, logger},
		diskExecutioner:    diskExecutioner{cfg.Disk, executor, logger},
		kernelExecutioner:  kernelExecutioner{cfg.Kernel, executor, logger},
	}
}

func (e *osExecutioner) Execute(ctx context.Context) error {
	if err := e.shellExecutioner.Process(ctx); err != nil {
		return fmt.Errorf("failed to execute shell instructions: %w", err)
	}

	if err := e.networkExecutioner.Process(ctx); err != nil {
		return fmt.Errorf("failed to execute network instructions: %w", err)
	}

	if err := e.diskExecutioner.Process(ctx); err != nil {
		return fmt.Errorf("failed to execute disk instructions: %w", err)
	}

	if err := e.kernelExecutioner.Process(ctx); err != nil {
		return fmt.Errorf("failed to execute kernel instructions: %w", err)
	}

	return nil
}

func (e *osExecutioner) Validate() error {
	return errors.Join(
		e.shellExecutioner.Validate(),
		e.networkExecutioner.Validate(),
		e.diskExecutioner.Validate(),
		e.kernelExecutioner.Validate(),
	)
}
