package command

import (
	"io"
)

type Command interface {
	Execute(args []string, stdout io.Writer) error
	Name() string
}
