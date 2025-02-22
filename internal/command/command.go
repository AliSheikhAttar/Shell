package command

import "errors"

var (
    ErrTooManyArgs = errors.New("too many arguments")
    ErrInvalidArgs = errors.New("invalid arguments")
)

// Command represents a shell command
type Command interface {
    Execute(args []string) error
    Name() string
}