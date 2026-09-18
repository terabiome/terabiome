package models

import (
	"errors"
	"strings"
)

type Command struct {
	Executable string
	Arguments  []string
}

func NewCommandFromString(input string) (*Command, error) {
	segments := strings.Split(strings.TrimSpace(input), " ")
	if len(segments) == 0 {
		return nil, errors.New("empty input")
	}

	return &Command{
		Executable: segments[0],
		Arguments:  segments[1:],
	}, nil
}

func (c Command) String() string {
	if len(c.Arguments) == 0 {
		return c.Executable
	}
	return c.Executable + " " + strings.Join(c.Arguments, " ")
}
