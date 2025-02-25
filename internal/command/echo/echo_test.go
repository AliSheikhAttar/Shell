package echo

import (
	"bytes"
	"testing"
)

func TestEchoCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedOutput string
		wantErr     bool // Echo command generally should not return errors in these basic tests
	}{
		{
			name:        "no arguments",
			args:        []string{},
			expectedOutput: "\n", // Echo with no args prints just a newline
			wantErr:     false,
		},
		{
			name:        "single argument - hello",
			args:        []string{"hello"},
			expectedOutput: "hello\n",
			wantErr:     false,
		},
		{
			name:        "multiple arguments - hello world",
			args:        []string{"hello", "world"},
			expectedOutput: "hello world\n",
			wantErr:     false,
		},
		{
			name:        "argument with spaces - quoted string",
			args:        []string{"hello world"}, // Arguments are already split by shell, so no quotes needed here for echo itself
			expectedOutput: "hello world\n",
			wantErr:     false,
		},
		{
			name:        "empty string argument - double quotes",
			args:        []string{""}, // Pass empty string as argument
			expectedOutput: "\n", // Echo of empty string prints a newline
			wantErr:     false,
		},
		{
			name:        "empty string argument - single quotes",
			args:        []string{}, // Pass empty string as argument
			expectedOutput: "\n", // Echo of empty string prints a newline
			wantErr:     false,
		},
		{
			name:        "arguments with special characters",
			args:        []string{"!@#$", "%^&*()", "+=", `{}`},
			expectedOutput: "!@#$ %^&*() +=" + " {}\n", // Joined with spaces
			wantErr:     false,
		},
		{
			name:        "mixed arguments - strings and special chars",
			args:        []string{"hello", "!@#$", "world"},
			expectedOutput: "hello !@#$ world\n",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewEchoCommand()

			// Verify command name (optional, but good practice)
			if got := cmd.Name(); got != "echo" {
				t.Errorf("EchoCommand.Name() = %v, want %v", got, "echo")
			}

			// Capture stdout
			stdout := &bytes.Buffer{}
			err := cmd.Execute(tt.args, stdout)

			// Check for unexpected error
			if tt.wantErr {
				t.Errorf("Test '%s' expected error, but got nil", tt.name) // In these tests, wantErr is always false, but kept for structure
			} else if err != nil {
				t.Errorf("Test '%s' unexpected error: %v", tt.name, err)
			}

			// Check output
			gotOutput := stdout.String()
			if gotOutput != tt.expectedOutput {
				// For better diff in test failures, especially with newlines.
				if len(gotOutput) != len(tt.expectedOutput) || gotOutput != tt.expectedOutput {
					t.Errorf("EchoCommand.Execute() output mismatch for test '%s':\n"+
						"Got:\n`%s`\nWant:\n`%s`", tt.name, gotOutput, tt.expectedOutput)
				}
			}
		})
	}
}