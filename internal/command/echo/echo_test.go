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
		wantErr     bool 
	}{
		{
			name:        "no arguments",
			args:        []string{},
			expectedOutput: "\n",
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
			args:        []string{"hello world"}, 
			expectedOutput: "hello world\n",
			wantErr:     false,
		},
		{
			name:        "empty string argument - double quotes",
			args:        []string{""},
			expectedOutput: "\n",
			wantErr:     false,
		},
		{
			name:        "empty string argument - single quotes",
			args:        []string{}, 
			expectedOutput: "\n", 
			wantErr:     false,
		},
		{
			name:        "arguments with special characters",
			args:        []string{"!@#$", "%^&*()", "+=", `{}`},
			expectedOutput: "!@#$ %^&*() +=" + " {}\n",
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

			if got := cmd.Name(); got != "echo" {
				t.Errorf("EchoCommand.Name() = %v, want %v", got, "echo")
			}

			stdout := &bytes.Buffer{}
			err := cmd.Execute(tt.args, stdout)

			if tt.wantErr {
				t.Errorf("Test '%s' expected error, but got nil", tt.name) // In these tests, wantErr is always false, but kept for structure
			} else if err != nil {
				t.Errorf("Test '%s' unexpected error: %v", tt.name, err)
			}

			// Check output
			gotOutput := stdout.String()
			if gotOutput != tt.expectedOutput {
				if len(gotOutput) != len(tt.expectedOutput) || gotOutput != tt.expectedOutput {
					t.Errorf("EchoCommand.Execute() output mismatch for test '%s':\n"+
						"Got:\n`%s`\nWant:\n`%s`", tt.name, gotOutput, tt.expectedOutput)
				}
			}
		})
	}
}