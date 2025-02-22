package shell

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

// mockShell creates a shell with custom input and output for testing
func mockShell(input string) (*Shell, *bytes.Buffer) {
	inputReader := strings.NewReader(input)
	outputBuffer := new(bytes.Buffer)

	shell := New()
	shell.reader = bufio.NewReader(inputReader)

	return shell, outputBuffer
}

func TestShellReadInput(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic input",
			input:    "hello\n",
			expected: "hello",
		},
		{
			name:     "input with spaces",
			input:    "  hello world  \n",
			expected: "hello world",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shell, _ := mockShell(tc.input)

			got, err := shell.readInput()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "empty input",
			input:    "",
			wantCmd:  "",
			wantArgs: nil,
		},
		{
			name:     "command only",
			input:    "exit",
			wantCmd:  "exit",
			wantArgs: []string{},
		},
		{
			name:     "command with one argument",
			input:    "exit 1",
			wantCmd:  "exit",
			wantArgs: []string{"1"},
		},
		{
			name:     "command with multiple arguments",
			input:    "exit 1 2 3",
			wantCmd:  "exit",
			wantArgs: []string{"1", "2", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			cmd, args := s.parseCommand(tt.input)

			if cmd != tt.wantCmd {
				t.Errorf("parseCommand() command = %v, want %v", cmd, tt.wantCmd)
			}

			if len(args) != len(tt.wantArgs) {
				t.Errorf("parseCommand() args length = %v, want %v", len(args), len(tt.wantArgs))
			}

			for i := range args {
				if args[i] != tt.wantArgs[i] {
					t.Errorf("parseCommand() arg[%d] = %v, want %v", i, args[i], tt.wantArgs[i])
				}
			}
		})
	}
}
