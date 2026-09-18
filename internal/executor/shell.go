package executor

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
)

type LocalShell struct {
	logger *slog.Logger
}

func NewLocalShell() *LocalShell {
	return &LocalShell{
		logger: slog.Default(),
	}
}

func (e *LocalShell) Name() string {
	return "local-shell"
}

func (e *LocalShell) Execute(ctx context.Context, input *Input) Output {
	output := Output{}

	cmdStr := input.CommandString()
	e.logger.Debug("executing command locally", slog.String("cmd", cmdStr))

	cmd := exec.CommandContext(ctx, input.Command.Executable, input.Command.Arguments...)
	cmd.Stdout = input.Stdout
	cmd.Stderr = input.Stderr

	output, err := NewOutput(input.Mode, cmd)
	if err != nil {
		output.fail(err)
		return output
	}

	e.execute(input, &output)
	return output
}

func (e *LocalShell) execute(input *Input, output *Output) error {
	cmd := output.Command

	if input.Mode == ModeAsync {
		if err := cmd.Start(); err != nil {
			output.fail(err)
			return fmt.Errorf("command failed to start: %w", err)
		}
		go e.handleResponse(input, output, cmd.Wait())
		return nil
	}

	return e.handleResponse(input, output, cmd.Run())
}

func (e *LocalShell) handleResponse(input *Input, output *Output, err error) error {
	cmdStr := input.CommandString()
	if err == nil {
		e.logger.Debug("command succeeded",
			slog.String("mode", string(input.Mode)),
			slog.String("cmd", cmdStr),
		)
		output.complete()
		return nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode := exitErr.ExitCode()
		e.logger.Warn("command failed",
			slog.String("mode", string(input.Mode)),
			slog.String("cmd", cmdStr),
			slog.Int("exit_code", exitCode),
		)

		output.fail(&CommandExitError{cmdStr, exitErr})
		return output.Error
	}

	e.logger.Error("command execution error",
		slog.String("mode", string(input.Mode)),
		slog.String("cmd", cmdStr),
		slog.String("error", err.Error()),
	)
	output.fail(&CommandExecutionError{cmdStr, err})
	return output.Error
}
