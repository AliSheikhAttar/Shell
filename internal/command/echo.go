package command

import (
	"fmt"
	"io"
	"strings"
)

type EchoCommand struct{}

func NewEchoCommand() *EchoCommand {
	return &EchoCommand{}
}

func (c *EchoCommand) Name() string {
	return "echo"
}

func (c *EchoCommand) Execute(args []string, stdout io.Writer) error {
	output := strings.Join(args, " ")
	fmt.Fprintln(stdout, output)
	return nil
}
