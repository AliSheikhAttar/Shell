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
	// Updated map to include usage for each command
	commands := map[string][2]string{
		
		"cd":      {"change your directory", "cd <path>"},
		"ls":      {"see the content of the directory", "ls [options]"},
		"cat":     {"see the content of the files", "cat <filename>"},
		"pwd":     {"current directory path", "pwd"},
		"type":    {"type of a command", "type <command>"},
		"adduser": {"register user to shell", "adduser {username} {password | empty}"},
		"echo":    {"write text/variables to output", "echo <text>"},
		"login":   {"login to shell as user", "login {username} {password | empty}"},
		"logout":  {"logout the shell", "logout"},
		"exit":    {"exit the shell", "exit [status code]"},
		"color":   {"set on/off color mode", "color [on|off]"},
		"history": {"history of executed commands", "history | history clean"},
	}

	fmt.Fprintln(stdout, "-------------------------------------------------------------------------------------------")
	// Adding the new "Usage" header
	coloredCommand := utils.ColorText("Command", utils.TextBlue, utils.Bold)
	coloredDescription := utils.ColorText("Description", utils.TextBlue, utils.Bold)
	coloredUsage := utils.ColorText("Usage", utils.TextBlue, utils.Bold)
	// Format now includes the "Usage" column
	fmt.Fprintf(stdout, "|   %s   |             %s           |           %s                       |\n", coloredCommand, coloredDescription, coloredUsage)
	fmt.Fprintln(stdout, "-------------------------------------------------------------------------------------------")
	for key, val := range commands {
		coloredKey := utils.ColorText(key, utils.TextYellow)
		// The third element (val[1]) is the "Usage" information
		fmt.Fprintf(stdout, "| %-20s | %-42s | %-46s |\n", coloredKey, utils.ColorText(val[0], utils.TextMagenta), utils.ColorText(val[1], utils.TextCyan))
	}
	fmt.Fprintln(stdout, "-------------------------------------------------------------------------------------------")
	return nil
}
