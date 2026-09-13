package executor

import (
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
	Command *exec.Cmd
	Signal  int
	done    chan bool

	Result any
	Error  error
}

func (o *Output) fail(err error) {
	o.done <- true
	o.Signal = -1
	o.Error = err
}

func (o *Output) complete() {
	o.done <- true
	o.Signal = 0
}

// expose read-only channel
func (o Output) Done() <-chan bool {
	return o.done
}
