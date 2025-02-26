package color

import (
	"asa/shell/utils"
	"fmt"
	"io"
	"os"
	"strings"
)

type ColorCommand struct {
	envVar string
}

func NewColorCommand() *ColorCommand {
	colorCmd := &ColorCommand{envVar: "SHELLCOLOR"}
	os.Setenv(colorCmd.envVar, "1")
	return colorCmd
}

func (c *ColorCommand) Name() string {
	return "color"
}

func (c *ColorCommand) Execute(args []string, stdout io.Writer) error {
	switch len(args) {
	case 0:
		return utils.ErrNotEnoughArgs
	case 1:

		if strings.ToLower(args[0]) == "off" {
			return c.unSet(stdout)
		}
		if strings.ToLower(args[0]) == "on" {
			return c.set(stdout)
		}
		return utils.ErrUnvalidArg
	default:
		return utils.ErrTooManyArgs
	}
}

func (c *ColorCommand) unSet(stdout io.Writer) error {

	if _, ok := os.LookupEnv(c.envVar); !ok {
		return utils.ErrColorUnset
	}
	os.Unsetenv(c.envVar)
	fmt.Fprintln(stdout, "Color is set off")
	return nil
}
func (c *ColorCommand) set(stdout io.Writer) error {
	if _, ok := os.LookupEnv(c.envVar); ok {
		return utils.ErrColorSet
	}
	os.Setenv(c.envVar, "1")

	fmt.Fprintln(stdout, utils.ColorText("Color is set on", utils.TextBlue))
	return nil
}
