package executioner

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/terabiome/infrastructures/internal/executor"
	"github.com/terabiome/infrastructures/internal/models"
)

type shellExecutioner struct {
	cfg      models.ShellConfig
	executor Executor
	logger   Logger
}

func (e *shellExecutioner) Validate() error {
	return errors.Join(
		e.cfg.SoftValidate(),
	)
}

func (e *shellExecutioner) Process(ctx context.Context) error {
	// ------------
	// --- SSHD ---
	// ------------

	// mutually exclusive
	if e.cfg.DisableService || e.cfg.EnableService {
		input := executor.Input{
			Mode:   executor.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: models.Command{
				Executable: "systemctl",
				Arguments:  []string{},
			},
		}
		if e.cfg.DisableService {
			input.Command.Arguments = append(input.Command.Arguments, "disable")
		} else {
			input.Command.Arguments = append(input.Command.Arguments, "enable")
		}
		input.Command.Arguments = append(input.Command.Arguments, "sshd")
		output := e.executor.Execute(ctx, &input)

		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to enable/disable sshd: %w", output.Error)
		}
	}

	if e.cfg.StartService {
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:   executor.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: models.Command{
				Executable: "systemctl",
				Arguments:  []string{"start", "sshd"},
			},
		})
		if output.Done() && output.Error != nil {
			return fmt.Errorf("failed to start sshd: %w", output.Error)
		}
	}

	return nil
}
