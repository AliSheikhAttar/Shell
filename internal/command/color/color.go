package color

import (
	"asa/shell/utils"
	"io"
	"os"
)

// CDCommand implements the 'cd' built-in command
type ColorCommand struct {
	envVar string
}

// NewCDCommand creates a new cd command
func NewColorCommand() *ColorCommand {
	colorCmd := &ColorCommand{envVar: "SHELLCOLOR"}
	os.Setenv(colorCmd.envVar, "1")
	return colorCmd
}

// Name returns the name of the command
func (c *ColorCommand) Name() string {
	return "color"
}

// Execute handles the cd command execution
func (c *ColorCommand) Execute(args []string, stdout io.Writer) error {
	switch len(args) {
	case 0:
		return utils.ErrNotEnoughArgs
	case 1:
		if args[0] == "off" {
			return c.unSet()
		}
		if args[0] == "on" {
			return c.set()
		}
		return utils.ErrUnvalidArg
	default:
		return utils.ErrTooManyArgs
	}
}

func (c *ColorCommand) unSet() error {

	if _, ok := os.LookupEnv(c.envVar); !ok {
		return utils.ErrColorUnset
	}
	os.Unsetenv(c.envVar)
	return nil
}
func (c *ColorCommand) set() error {

	err := os.Setenv(c.envVar, "1")
	if err != nil {
		return utils.ErrColorWrong
	}
	return nil
}
