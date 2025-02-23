package shell

import (
	"asa/shell/internal/command"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
			cmd, args, _, _ := s.parseCommand(tt.input)

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

func TestSystemCommandExecution(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create a test bash script
	shScriptPath := filepath.Join(tmpDir, "test-script.sh")
	shScriptContent := `#!/bin/sh
		echo "Hello from test script"`
	if err := os.WriteFile(shScriptPath, []byte(shScriptContent), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a test python script
	pyScriptPath := filepath.Join(tmpDir, "test-script.py")
	pyScriptContent := `print("Hello from test script")`
	if err := os.WriteFile(pyScriptPath, []byte(pyScriptContent), 0755); err != nil {
		t.Fatal(err)
	}

	// Add temporary directory to PATH
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+":"+originalPath)
	defer os.Setenv("PATH", originalPath)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "execute built-in command",
			input:   "echo hello",
			wantErr: false,
		},
		{
			name:    "execute system command",
			input:   "test-script.sh",
			wantErr: false,
		},
		{
			name:    "execute system command with arguments",
			input:   fmt.Sprintf("python %s/test-script.py", tmpDir),
			wantErr: false,
		},
		{
			name:    "execute non-existent command",
			input:   "nonexistentcommand",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := New()
			_, err := sh.executeCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Shell.executeCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestShell_SystemCommandExecution(t *testing.T) {
	// Create temporary test directory and scripts
	tmpDir := t.TempDir()
	setupTestScripts(t, tmpDir)

	// Add temporary directory to PATH
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+":"+originalPath)
	defer os.Setenv("PATH", originalPath)

	tests := []struct {
		name    string
		input   string
		wantOut string
		wantErr bool
	}{
		{
			name:    "execute simple script",
			input:   "test-script.sh",
			wantOut: "Hello from test script\n",
			wantErr: false,
		},
		{
			name:    "execute script with arguments",
			input:   "echo-args.sh arg1 arg2",
			wantOut: "Arguments: arg1 arg2\n",
			wantErr: false,
		},
		{
			name:    "execute failing script",
			input:   "fail-script.sh",
			wantOut: "This script fails\n",
			wantErr: true,
		},
		{
			name:    "execute script with env variables",
			input:   "env-script.sh",
			wantOut: "TEST_VAR=test_value\n",
			wantErr: false,
		},
		{
			name:    "execute nonexistent command",
			input:   "nonexistent-command",
			wantOut: "",
			wantErr: true,
		},
		{
			name:    "execute command with full path",
			input:   "/usr/bin/echo test",
			wantOut: "test\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create shell instance with captured output
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			sh := createTestShell(stdout, stderr)

			// Set test environment variable if needed
			if strings.Contains(tt.input, "env-script.sh") {
				os.Setenv("TEST_VAR", "test_value")
				defer os.Unsetenv("TEST_VAR")
			}

			// Execute command
			_, err := sh.executeCommand(tt.input)

			// Check error condition
			if (err != nil) != tt.wantErr {
				t.Errorf("Shell.executeCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check output
			if tt.wantOut != "" {
				gotOut := stdout.String()
				if gotOut != tt.wantOut {
					t.Errorf("Shell.executeCommand() output = %q, want %q", gotOut, tt.wantOut)
				}
			}
		})
	}
}
func TestShell_CommandWithPipes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		stdin   string
		wantOut string
		wantErr bool
	}{
		{
			name:    "cat command with stdin",
			input:   "cat",
			stdin:   "test demo\n",
			wantOut: "test demo\n",
			wantErr: false,
		},
		{
			name:    "cat command with multiple lines",
			input:   "cat",
			stdin:   "line1\nline2\n",
			wantOut: "line1\nline2\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stdin := strings.NewReader(tt.stdin)
			sh := createTestShellWithStdin(stdin, stdout, nil)

			err := sh.executeSystemCommand(tt.input, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Shell.executeCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := stdout.String()
			if got != tt.wantOut {
				t.Errorf("Shell.executeCommand() output = %q, want %q", got, tt.wantOut)
			}
		})
	}
}

func TestShell_ExecutablePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a non-executable script
	nonExecPath := filepath.Join(tmpDir, "non-executable.sh")
	err := os.WriteFile(nonExecPath, []byte("#!/bin/sh\necho test\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test executing non-executable file
	stdout := &bytes.Buffer{}
	sh := createTestShell(stdout, nil)
	_, err = sh.executeCommand(nonExecPath)
	if err == nil {
		t.Error("Shell.executeCommand() should fail for non-executable file")
	}
}

// Helper functions

func setupTestScripts(t *testing.T, tmpDir string) {
	scripts := map[string]string{
		"test-script.sh": `#!/bin/sh
echo "Hello from test script"
`,
		"echo-args.sh": `#!/bin/sh
echo "Arguments: $@"
`,
		"fail-script.sh": `#!/bin/sh
echo "This script fails"
exit 1
`,
		"env-script.sh": `#!/bin/sh
echo "TEST_VAR=$TEST_VAR"
`,
	}

	for name, content := range scripts {
		path := filepath.Join(tmpDir, name)
		err := os.WriteFile(path, []byte(content), 0755)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func createTestShell(stdout, stderr io.Writer) *Shell {
	sh := New()
	// Set custom stdout/stderr for testing
	if stdout != nil {
		sh.stdout = stdout
	}
	if stderr != nil {
		sh.stderr = stderr
	}
	return sh
}
func createTestShellWithStdin(stdin io.Reader, stdout, stderr io.Writer) *Shell {
	sh := New()
	if stdin != nil {
		sh.stdin = stdin
		sh.reader = bufio.NewReader(stdin) // Make sure to update the reader too
	}
	if stdout != nil {
		sh.stdout = stdout
	}
	if stderr != nil {
		sh.stderr = stderr
	}
	return sh
}

// TestHelperProcess isn't a real test - it's used as a helper process for mocking commands
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		os.Exit(1)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "echo":
		fmt.Println(strings.Join(args, " "))
	case "cat":
		io.Copy(os.Stdout, os.Stdin)
	default:
		os.Exit(1)
	}
}

// shell/shell_test.go

func TestShell_PwdCommand(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "shell-pwd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	stdout := &bytes.Buffer{}
	sh := createTestShellWithStdin(nil, stdout, nil)
	sh.registerCommand(command.NewPwdCommand())
	_, err = sh.executeCommand("pwd")
	if err != nil {
		t.Errorf("Shell.executeCommand() error = %v", err)
		return
	}

	got := strings.TrimSpace(stdout.String())
	if got != tmpDir {
		t.Errorf("pwd command output = %v, want %v", got, tmpDir)
	}
}
