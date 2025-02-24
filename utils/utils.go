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
	ErrInvalidQuotedArg     = errors.New("invalid Quoted arg")
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

// isQuoted checks if a string is wrapped in quotes
func IsQuoted(s string) bool {
	return (HasPrefix(s, "'") && HasSuffix(s, "'")) ||
		(HasPrefix(s, "\"") && HasSuffix(s, "\""))
}

// isQuoted checks if a string is wrapped in quotes
func IsQuoted1(s string) bool {
	return len(s) != 1 && ((HasPrefix(s, "'") && HasSuffix(s, "'")) ||
		(HasPrefix(s, "\"") && HasSuffix(s, "\"")))
}

// isAlphaNumeric checks if a character is alphanumeric
func IsAlphaNumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9')
}

func HasPrefix(s string, prefix string) bool {
	w := !(s[0:len(prefix)] == prefix)
	return len(s) >= len(prefix) && !w
}

func HasSuffix(s string, suffix string) bool {
	w := !(s[len(s)-len(suffix):] == suffix)
	return len(s) >= len(suffix) && !w
}

func TrimEdge(s string) string {
	if len(s) < 1 {
		return s
	}
	return s[1:len(s)-1]
}
func ExtractQuotes(input string) ([]string, error) {
	res := []string{}
	for i := 0; i < len(input); i++ {
		if input[i] == '"' {
			var result strings.Builder
			result.WriteByte(input[i])
			for i = i + 1; i < len(input) && input[i] != '"'; i++ {
				result.WriteByte(input[i])
			}
			if i == len(input) || input[i] != '"' {
				result.WriteByte('"') // last quote
				res = append(res, result.String())
				return res, ErrInvalidQuotedArg
			}
			result.WriteByte(input[i]) // last quote
			res = append(res, result.String())
		} else if input[i] == '\'' {
			var result strings.Builder
			result.WriteByte(input[i])
			for i = i + 1; i < len(input) && input[i] != '\''; i++ {
				result.WriteByte(input[i])
			}
			if i == len(input) || input[i] != '\'' {
				result.WriteByte('\'') // last quote
				res = append(res, result.String())
				return res, ErrInvalidQuotedArg
			}
			result.WriteByte(input[i]) // last quote
			res = append(res, result.String())
		}
	}
	return res, nil
}

func Seperate(input string, quotes []string) (res []string) {
	fields := strings.Fields(input)
	j := 0
	for i := 0; i < len(fields); i++ {
		if HasPrefix(fields[i], string("'")) {
			if IsQuoted(fields[i]) && len(fields[i]) != 1 {
				res = append(res, quotes[j])
				j++
				continue
			}
			i++
			for i < len(fields) && !HasSuffix(fields[i], string("'")) {
				i++
			}
			res = append(res, quotes[j])
			j++
			continue
		} else if HasPrefix(fields[i], string('"')) {
			if IsQuoted(fields[i]) && len(fields[i]) != 1 {
				res = append(res, quotes[j])
				j++
				continue
			}
			i++
			for i < len(fields) && !HasSuffix(fields[i], string('"')) {
				i++
			}
			res = append(res, quotes[j])
			j++
			continue
		}
		res = append(res, fields[i])
	}
	return res
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
