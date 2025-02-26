package help

import (
	"asa/shell/utils"
	"fmt"
	"io"
)

type HelpCommand struct {
}

func NewHelpCommand() *HelpCommand {
	return &HelpCommand{}
}

func (c *HelpCommand) Name() string {
	return "--help"
}

func (c *HelpCommand) Execute(args []string, stdout io.Writer) error {
	if len(args) > 0 {
		return utils.ErrInvalidArgs
	}
	commands := map[string]string{
		"color":   "set on/off color mode",
		"cat":     "see the content of the files",
		"ls":      "see the content of the directory",
		"cd":      "change your directory",
		"echo":    "write text/variables to output",
		"type":    "type of a command",
		"pwd":     "current directory path",
		"exit":    "exit the shell",
		"history": "history of executed commands",
		"adduser": "register user to shell",
		"login":   "login to shell as user",
		"logout":  "logout the shell",
	}
	fmt.Fprintln(stdout, "---------------------------------------------------")
	coloredCommand := utils.ColorText("Command", utils.TextBlue, utils.Bold)
	coloredDescription := utils.ColorText("Description", utils.TextBlue, utils.Bold)
	fmt.Fprintf(stdout, "|   %s   |             %s           |\n", coloredCommand, coloredDescription)
	fmt.Fprintln(stdout, "---------------------------------------------------")
	for key, val := range commands {
		coloredKey := utils.ColorText(key, utils.TextYellow)
		fmt.Fprintf(stdout, "| %-20s | %-42s |\n", coloredKey, utils.ColorText(val, utils.TextMagenta))
	}
	fmt.Fprintln(stdout, "---------------------------------------------------")
	return nil
}
