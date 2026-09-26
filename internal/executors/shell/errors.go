package shellexecutors

import (
	"fmt"
	"os/exec"
)

type CommandExitError struct {
	CmdStr       string
	WrappedError *exec.ExitError
}

func (e *CommandExitError) Error() string {
	return fmt.Sprintf("command exited with code %d", e.WrappedError.ExitCode())
}

type CommandExecutionError struct {
	CmdStr       string
	WrappedError error
}

func (e *CommandExecutionError) Error() string {
	return fmt.Sprintf("command execution failed: %v", e.WrappedError)
}
