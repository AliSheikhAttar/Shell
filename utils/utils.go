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
	ErrPwdWentWrong         = errors.New("something went wrong while trying to identify current dir")
)

var LinuxBuiltins = map[string]bool{
	"cd":      true,
	"pwd":     true,
	"exit":    true,
	"echo":    true,
	"export":  true,
	"source":  true,
	"alias":   true,
	"unalias": true,
	"set":     true,
	"unset":   true,
	"exec":    true,
	"command": true,
	".":       true,
}

const (
	// Text colors
	TextBlack   = "\033[30m"
	TextRed     = "\033[31m"
	TextGreen   = "\033[32m"
	TextYellow  = "\033[33m"
	TextBlue    = "\033[34m"
	TextMagenta = "\033[35m"
	TextCyan    = "\033[36m"
	TextWhite   = "\033[37m"

	// Background colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"

	// Special formatting
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Blink     = "\033[5m"
	Reverse   = "\033[7m"
)

func ColorText(text string, formats ...string) string {
	var combined string
	for _, format := range formats {
		combined += format
	}
	return combined + text + Reset
}

func IsColor() bool {
	_, ok := os.LookupEnv("SHELLCOLOR")
	return ok
}

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
