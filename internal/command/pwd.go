package command

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "runtime"
)

type PwdCommand struct {
    stdout io.Writer
}

func NewPwdCommand(stdout io.Writer) *PwdCommand {
    return &PwdCommand{
        stdout: stdout,
    }
}

func (c *PwdCommand) Execute(args []string) error {
    pwd, err := c.getCurrentDirectory()
    if err != nil {
        return fmt.Errorf("failed to get current directory: %v", err)
    }

    _, err = fmt.Fprintln(c.stdout, pwd)
    return err
}

func (c *PwdCommand) getCurrentDirectory() (string, error) {
    // Try different methods to get the current directory
    
    // 1. Try PWD environment variable first
    if pwd := os.Getenv("PWD"); pwd != "" {
        if isValidDirectory(pwd) {
            return filepath.Clean(pwd), nil
        }
    }

    // 2. Try /proc/self/cwd on Linux
    if runtime.GOOS == "linux" {
        if pwd, err := os.Readlink("/proc/self/cwd"); err == nil {
            if isValidDirectory(pwd) {
                return filepath.Clean(pwd), nil
            }
        }
    }

    // 3. Try to resolve using filepath operations
    pwd, err := filepath.Abs(".")
    if err != nil {
        return "", err
    }

    // Verify the directory exists and is accessible
    if !isValidDirectory(pwd) {
        return "", fmt.Errorf("directory not accessible: %s", pwd)
    }

    return filepath.Clean(pwd), nil
}

func isValidDirectory(path string) bool {
    info, err := os.Stat(path)
    if err != nil {
        return false
    }
    return info.IsDir()
}

func (c *PwdCommand) Name() string {
    return "pwd"
}