package echo

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestEchoCommand(t *testing.T) {
	// Set up test environment variable
	os.Setenv("PATH1", "/usr/local/bin:/usr/bin:/bin")
	os.Setenv("HOME1", "/home/user")

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "simple text",
			args:     []string{"hello"},
			expected: "hello",
		},
		{
			name:     "multiple words",
			args:     []string{"hello", "world"},
			expected: "hello world",
		},
		{
			name:     "with environment variable",
			args:     []string{"Path1:", "$PATH1"},
			expected: "Path1: /usr/local/bin:/usr/bin:/bin",
		},
		{
			name:     "quoted text",
			args:     []string{"'hello PATH'"},
			expected: "hello PATH",
		},
		{
			name:     "quoted text with special char",
			args:     []string{"'hello $PATH'"},
			expected: "hello $PATH",
		},
		{
			name:     "mixed content",
			args:     []string{"Home1", "is", "$HOME"},
			expected: "Home1 is /home/asa",
		},
		{
			name:     "environment variable in word",
			args:     []string{"no$PATH1"},
			expected: "no/usr/local/bin:/usr/bin:/bin",
		},
		{
			name:     "quoted environment variable",
			args:     []string{"$'PATH'"},
			expected: "PATH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewEchoCommand()

			// Verify command name
			if got := cmd.Name(); got != "echo" {
				t.Errorf("EchoCommand.Name() = %v, want %v", got, "echo")
			}

			// Capture stdout
			// Note: In a real implementation, you might want to use a more sophisticated
			// way to capture stdout, but this is simplified for the example
			stdout := &bytes.Buffer{}
			err := cmd.Execute(tt.args, stdout)
			if err != nil {
				t.Errorf("EchoCommand.Execute() error = %v", err)
			}
			got := strings.TrimSpace(stdout.String())
			if got != tt.expected {
				t.Errorf("PwdCommand.Execute() output = %v, want %v", got, tt.expected)
			}
		})
	}
}
