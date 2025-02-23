package command

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

// ExitCommand implements the 'exit' built-in command
type ExitCommand struct {}

// NewExitCommand creates a new exit command
func NewExitCommand() *ExitCommand {
	return &ExitCommand{}
}

// Name returns the name of the command
func (c *ExitCommand) Name() string {
	return "exit"
}

// Execute handles the exit command execution
func (c *ExitCommand) Execute(args []string, stdout io.Writer) error {
	switch len(args) {
	case 0:
		fmt.Fprintln(stdout, "exit status 0")
		os.Exit(0)

	case 1:
		status, err := strconv.Atoi(args[0])
		if err != nil {
			return ErrInvalidArgs
		}
		fmt.Fprintln(stdout, "exit status ", status)
		os.Exit(0)
	default:
		return ErrTooManyArgs
	}
	return nil // This line will never be reached due to os.Exit
}
