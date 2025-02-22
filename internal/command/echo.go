package command

import (
	"asa/shell/utils"
	"fmt"
	"io"
	"os"
	"strings"
)

type EchoCommand struct{
	stdout io.Writer
}

func NewEchoCommand(stdout io.Writer) *EchoCommand {
	return &EchoCommand{
		stdout: stdout,
	}
}

func (c *EchoCommand) Name() string {
	return "echo"
}

func (c *EchoCommand) Execute(args []string) error {
	output := expandArgs(args)
	fmt.Fprintln(c.stdout, output)
	return nil
}

// expandArgs processes the arguments and expands environment variables
func expandArgs(args []string) string {
	var result []string

	for _, arg := range args {
		// Check if the argument is a quoted string
		if isQuoted(arg) {
			// Remove quotes and add as literal
			result = append(result, strings.Trim(arg, "'\""))
			continue
		}

		// Expand environment variables
		expanded := expandEnvVars(arg)
		result = append(result, expanded)
	}

	return strings.Join(result, " ")
}

// expandEnvVars replaces environment variables with their values
func expandEnvVars(s string) string {
	var result strings.Builder
	i := 0

	for i < len(s) {
		if s[i] == '$' && i+1 < len(s) {
			j := i + 1
			// Find the end of the variable name
			for j < len(s) && (isAlphaNumeric(s[j]) || s[j] == '_') {
				j++
			}

			if j > i+1 {
				varName := s[i+1 : j]
				varValue := os.Getenv(varName)
				result.WriteString(varValue)
				i = j
				continue
			}

			// cases like $"path"
			if utils.HasPrefix(string(s[j]), "'") {
				j++
				k := j + 1
				for !utils.HasSuffix(s[j:k], "'") {
					k++
				}

				result.WriteString(s[j : k-1])
				i = k
				continue
			}
			if utils.HasPrefix(string(s[j]), "\"") {
				j++
				k := j + 1
				for !utils.HasSuffix(s[j:k], "\"") {
					k++
				}

				result.WriteString(s[j : k-1])
				i = k
				continue
			}
		}
		result.WriteByte(s[i])
		i++
	}

	return result.String()
}

// isQuoted checks if a string is wrapped in quotes
func isQuoted(s string) bool {
	return (utils.HasPrefix(s, "'") && utils.HasSuffix(s, "'")) ||
		(utils.HasPrefix(s, "\"") && utils.HasSuffix(s, "\""))
}

// isAlphaNumeric checks if a character is alphanumeric
func isAlphaNumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9')
}

