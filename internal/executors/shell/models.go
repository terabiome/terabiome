package shellexecutors

import (
	"errors"
	"io"
	"os/exec"
	"strings"
)

type execMode string

const (
	ModeSync  execMode = "sync"
	ModeAsync execMode = "async"
)

type CommandInput struct {
	Mode    execMode
	Stdout  io.Writer
	Stderr  io.Writer
	Command Command
}

func (i CommandInput) CommandString() string {
	return i.Command.String()
}

type Command struct {
	Executable string
	Arguments  []string
}

func (c Command) String() string {
	if len(c.Arguments) == 0 {
		return c.Executable
	}
	return c.Executable + " " + strings.Join(c.Arguments, " ")
}

type CommandOutput struct {
	mode execMode

	Command *exec.Cmd
	Signal  int
	done    chan struct{}

	Result any
	Error  error
}

func NewOutput(mode execMode, cmd *exec.Cmd) (CommandOutput, error) {
	if cmd == nil {
		return CommandOutput{}, errors.New("command is nil")
	}
	return CommandOutput{
		mode:    mode,
		Command: cmd,
		done:    make(chan struct{}, 1),
	}, nil
}

func (o CommandOutput) IsAsync() bool {
	return o.mode == ModeAsync
}

func (o CommandOutput) Done() bool {
	return len(o.done) != 0
}

func (o *CommandOutput) Wait() {
	<-o.done
}

func (o *CommandOutput) fail(err error) {
	o.done <- struct{}{}
	o.Signal = -1
	o.Error = err
}

func (o *CommandOutput) complete() {
	o.done <- struct{}{}
	o.Signal = 0
}
