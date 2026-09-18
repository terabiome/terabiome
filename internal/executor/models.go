package executor

import (
	"errors"
	"io"
	"os/exec"

	"github.com/terabiome/infrastructures/internal/models"
)

type execMode string

const (
	ModeSync  execMode = "sync"
	ModeAsync execMode = "async"
)

type Input struct {
	Mode    execMode
	Stdout  io.Writer
	Stderr  io.Writer
	Command models.Command
}

func (i Input) CommandString() string {
	return i.Command.String()
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
