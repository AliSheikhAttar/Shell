package command

import (
	"errors"
	"io"
)

var (
    ErrTooManyArgs = errors.New("too many arguments")
    ErrInvalidArgs = errors.New("invalid arguments")
)

// Command represents a shell command
type Command interface {
    Execute(args []string, stdout io.Writer) error
    Name() string
}