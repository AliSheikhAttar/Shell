package cd

import (
	"asa/shell/utils"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNoFileDir = errors.New("no such file or directory")
)

// CDCommand implements the 'cd' built-in command
type CDCommand struct {
	rootDir string
}

// NewCDCommand creates a new cd command
func NewCDCommand(rootDir string) *CDCommand {
	return &CDCommand{
		rootDir: rootDir,
	}
}

// Name returns the name of the command
func (c *CDCommand) Name() string {
	return "cd"
}

// Execute handles the cd command execution
func (c *CDCommand) Execute(args []string, stdout io.Writer) error {
	var dir string

	switch len(args) {
	case 0:
		// cd without arguments - go to home directory
		dir = c.rootDir
	case 1:
		// cd with one argument
		switch args[0] {
		case "~":
			// cd ~ - go to home directory
			dir = os.Getenv("HOME")
		default:
			// cd <path> - go to specified directory
			dir = args[0]
			// Handle relative paths that start with ~
			if len(dir) > 0 && dir[0] == '~' {
				home := os.Getenv("HOME")
				dir = filepath.Join(home, dir[1:])
				break
			}
			if dir == ".." {
				currentDir, err := utils.CurrentPwd()
				if err != nil {
					return err
				}
				currentCleanDir := filepath.Clean(currentDir)
				dirArgs := strings.Split(currentCleanDir, "/")
				dirArgs = dirArgs[:len(dirArgs)-1]
				newDir := strings.Join(dirArgs, "/")
				if utils.IsValidDirectory(newDir) {
					dir = newDir
					break

				}
			}
		}
	default:
		return utils.ErrTooManyArgs
	}

	// Try to change directory
	if err := os.Chdir(dir); err != nil {
		return ErrNoFileDir
	}

	return nil
}
