package command

import (
	"io"
)

// Command represents a shell command
type Command interface {
	Execute(args []string, stdout io.Writer) error
	Name() string
}
