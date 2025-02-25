package command

import (
	"asa/shell/utils"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	// "runtime"
)

type PwdCommand struct{}

func NewPwdCommand() *PwdCommand {
	return &PwdCommand{}
}

func (c *PwdCommand) Execute(args []string, stdout io.Writer) error {
	pwd, err := c.getCurrentDirectory()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	_, err = fmt.Fprintln(stdout, pwd)
	return err
}

func (c *PwdCommand) getCurrentDirectory() (string, error) {

	if pwd, err := filepath.Abs("."); pwd != "" && err == nil {
		if utils.IsValidDirectory(pwd) {
			return filepath.Clean(pwd), nil
		}
	}
	// Try PWD environment variable first
	if pwd := os.Getenv("PWD"); pwd != "" {
		if utils.IsValidDirectory(pwd) {
			return filepath.Clean(pwd), nil
		}
	}

	// Try /proc/self/cwd on Linux
	if runtime.GOOS == "linux" {
		if pwd, err := os.Readlink("/proc/self/cwd"); err == nil {
			if utils.IsValidDirectory(pwd) {
				return filepath.Clean(pwd), nil
			}
		}
	}
	return "", nil
}

func (c *PwdCommand) Name() string {
	return "pwd"
}
