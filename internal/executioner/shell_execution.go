package executioner

import (
	"context"
	"errors"
	"fmt"
	"os"

	yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"
	shellexecutors "github.com/terabiome/terabiome/internal/executors/shell"
)

type shellExecutioner struct {
	cfg      yamlcontracts.ShellConfig
	executor CommandExecutor
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
		input := shellexecutors.CommandInput{
			Mode:   shellexecutors.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: shellexecutors.Command{
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
		output := e.executor.Execute(ctx, &shellexecutors.CommandInput{
			Mode:   shellexecutors.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: shellexecutors.Command{
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
