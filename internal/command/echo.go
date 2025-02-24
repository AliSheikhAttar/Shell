package command

import (
	"asa/shell/utils"
	"fmt"
	"io"
	"os"
	"strings"
)

type EchoCommand struct{}

func NewEchoCommand() *EchoCommand {
	return &EchoCommand{}
}

func (c *EchoCommand) Name() string {
	return "echo"
}

func (c *EchoCommand) Execute(args []string, stdout io.Writer) error {
	output := expandArgs(args)
	fmt.Fprintln(stdout, output)
	return nil
}

// expandArgs processes the arguments and expands environment variables
func expandArgs(args []string) string {
	var result []string

	for _, arg := range args {
		// Check if the argument is a quoted string
		if utils.IsQuoted(arg) {
			// Remove quotes and add as literal
			result = append(result, utils.TrimEdge(arg))
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
			for j < len(s) && (utils.IsAlphaNumeric(s[j]) || s[j] == '_') {
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
