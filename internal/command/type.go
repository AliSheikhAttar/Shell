package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
    ErrCommandNotFound = errors.New("command not found")
    ErrMissingCommandName = errors.New("type: missing command name")
    ErrEnvironmentVarNotSet = errors.New("PATH environment variable is not set")
)
type TypeCommand struct {
    builtins map[string]bool // Map of built-in commands
    stdout io.Writer
}

func NewTypeCommand(builtins []string, stdout io.Writer) *TypeCommand {
    // Create a map of built-in commands for efficient lookup
    builtinsMap := make(map[string]bool)
    for _, cmd := range builtins {
        builtinsMap[cmd] = true
    }
    
    return &TypeCommand{
        builtins: builtinsMap,
        stdout : stdout,
    }
}

func (c *TypeCommand) Name() string {
    return "type"
}

func (c *TypeCommand) Execute(args []string) error {
    if len(args) == 0 {
        return ErrMissingCommandName
    }

    for _, cmd := range args {
        if err := c.findCommand(cmd); err != nil {
            return err
        }
    }

    return nil
}

func (c *TypeCommand) findCommand(cmd string) error {
    // Check if it's a built-in command
    if c.builtins[cmd] {
        fmt.Printf("%s is a shell builtin\n", cmd)
        return nil
    }

    // If not built-in, search in PATH
    path := os.Getenv("PATH")
    if path == "" {
        return ErrEnvironmentVarNotSet
    }

    // Search in each directory in PATH
    dirs := strings.Split(path, ":")
    for _, dir := range dirs {
        fullPath := filepath.Join(dir, cmd)
        if fileInfo, err := os.Stat(fullPath); err == nil {
            // Check if the file is executable
            if fileInfo.Mode()&0111 != 0 {
                fmt.Printf("%s is %s\n", cmd, fullPath)
                return nil
            }
        }
    }

    return ErrCommandNotFound
}