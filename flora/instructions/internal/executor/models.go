package executor

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

type Input struct {
	Mode       execMode
	Stdout     io.Writer
	Stderr     io.Writer
	Executable string
	Arguments  []string
}

func (i Input) CommandString() string {
	if len(i.Arguments) == 0 {
		return i.Executable
	}
	return i.Executable + " " + strings.Join(i.Arguments, " ")
}

type Output struct {
	mode execMode

	Command *exec.Cmd
	Signal  int
	done    chan struct{}

	Result any
	Error  error
}

func NewOutput(mode execMode, cmd *exec.Cmd) (Output, error) {
	if cmd == nil {
		return Output{}, errors.New("command is nil")
	}
	return Output{
		mode:    mode,
		Command: cmd,
		done:    make(chan struct{}, 1),
	}, nil
}

func (o Output) IsAsync() bool {
	return o.mode == ModeAsync
}

func (o Output) Done() bool {
	return len(o.done) != 0
}

func (o *Output) Wait() {
	<-o.done
}

func (o *Output) fail(err error) {
	o.done <- struct{}{}
	o.Signal = -1
	o.Error = err
}

func (o *Output) complete() {
	o.done <- struct{}{}
	o.Signal = 0
}
