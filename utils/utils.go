package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrCommandNotFound      = errors.New("command not found")
	ErrEnvironmentVarNotSet = errors.New("PATH environment variable is not set")
	ErrPwdWentWrong         = errors.New("Something went wrong while trying to identify current dir")
)

func FindCommand(cmd string) (string, error) {
	// Check if it's a built-in command
	builtins := []string{"exit", "echo", "cat", "type", "cd"} // Add all built-in commands here
	for _, builtin := range builtins {
		if builtin == cmd {
			return fmt.Sprintf("$builtin:%s", builtin), nil
		}
	}
	// if executable file
	if strings.Contains(cmd, "/") {
		return cmd, nil // exec files with or without suffix
	}

	// If not built-in, search in PATH
	path := os.Getenv("PATH")
	if path == "" {
		return "", ErrEnvironmentVarNotSet
	}

	// Search in each directory in PATH
	dirs := strings.Split(path, ":")
	for _, dir := range dirs {
		fullPath := filepath.Join(dir, cmd)
		if isExecutable(fullPath) {
			return fullPath, nil
		}
	}

	return "", ErrCommandNotFound
}

func HasPrefix(s string, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}

func HasSuffix(s string, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func HandleAdress(baseAddr string, currentDir string) string {
	cleanBaseAddr := strings.Trim(baseAddr, "\n")
	cleanCurrentDir := strings.Trim(currentDir, "\n")

	if HasPrefix(cleanCurrentDir, cleanBaseAddr) {
		if cleanBaseAddr == cleanCurrentDir {
			return "~"
		}
		return "~" + cleanCurrentDir[len(cleanBaseAddr):]
	}
	return cleanCurrentDir
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && (info.Mode()&0111 != 0)
}

func IsValidDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func CurrentPwd() (string, error) {
	if pwd, err := filepath.Abs("."); pwd != "" && err == nil {
		if IsValidDirectory(pwd) {
			return filepath.Clean(pwd), nil
		}
	}
	return "", ErrPwdWentWrong
}
