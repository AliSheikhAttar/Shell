package typecmd

import (
	"asa/shell/utils"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type TypeCommand struct {
	builtins map[string]bool 
}

func NewTypeCommand(builtins []string) *TypeCommand {
	builtinsMap := make(map[string]bool)
	for _, cmd := range builtins {
		builtinsMap[cmd] = true
	}

	return &TypeCommand{
		builtins: builtinsMap,
	}
}

func (c *TypeCommand) Name() string {
	return "type"
}

func (c *TypeCommand) Execute(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return utils.ErrMissingCommandName
	}
	for _, cmd := range args {
		result, err := c.findCommand(cmd)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, result)
	}

	return nil
}

func (c *TypeCommand) findCommand(cmd string) (string, error) {
	if c.builtins[cmd] {
		return fmt.Sprintf("%s is a shell builtin", cmd), nil
	}

	path := os.Getenv("PATH")
	if path == "" {
		return "", utils.ErrEnvironmentVarNotSet
	}

	dirs := strings.Split(path, ":")
	for _, dir := range dirs {
		fullPath := filepath.Join(dir, cmd)
		if fileInfo, err := os.Stat(fullPath); err == nil {
			if fileInfo.Mode()&0111 != 0 {

				return fmt.Sprintf("%s is %s", cmd, fullPath), nil
			}
		}
	}

	return "", utils.ErrCommandNotFound
}
