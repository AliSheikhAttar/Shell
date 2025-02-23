package command

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)
var (
    ErrNoFileDir = errors.New("no such file or directory")
)

// CDCommand implements the 'cd' built-in command
type CDCommand struct{}

// NewCDCommand creates a new cd command
func NewCDCommand() *CDCommand {
    return &CDCommand{}
}

// Name returns the name of the command
func (c *CDCommand) Name() string {
    return "cd"
}

// Execute handles the cd command execution
func (c *CDCommand) Execute(args []string, stdout io.Writer) error {
    var dir string
    var err error

    switch len(args) {
    case 0:
        // cd without arguments - go to home directory
        dir, err = os.UserHomeDir()
        if err != nil {
            return err
        }
    case 1:
        // cd with one argument
        switch args[0] {
        case "~":
            // cd ~ - go to home directory
            dir, err = os.UserHomeDir()
            if err != nil {
                return err
            }
        default:
            // cd <path> - go to specified directory
            dir = args[0]
            // Handle relative paths that start with ~
            if len(dir) > 0 && dir[0] == '~' {
                home, err := os.UserHomeDir()
                if err != nil {
                    return err
                }
                dir = filepath.Join(home, dir[1:])
            }
        }
    default:
        return ErrTooManyArgs
    }

    // Try to change directory
    if err := os.Chdir(dir); err != nil {
        return ErrNoFileDir
    }

    return nil
}