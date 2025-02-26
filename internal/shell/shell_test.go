package shell

import (
	"asa/shell/internal/command"
	"asa/shell/internal/command/adduser"
	"asa/shell/internal/command/cat"
	"asa/shell/internal/command/cd"
	"asa/shell/internal/command/color"
	"asa/shell/internal/command/echo"
	"asa/shell/internal/command/exit"
	"asa/shell/internal/command/help"
	"asa/shell/internal/command/history"
	"asa/shell/internal/command/login"
	"asa/shell/internal/command/logout"
	"asa/shell/internal/command/ls"
	"asa/shell/internal/command/pwd"
	typecmd "asa/shell/internal/command/type"
	db "asa/shell/internal/database"
	"asa/shell/internal/redirection"
	user "asa/shell/internal/service"
	"asa/shell/utils"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	tmpDir := t.TempDir()

	shScriptPath := filepath.Join(tmpDir, "test-script.sh")
	shScriptContent := `#!/bin/sh
		echo "Hello from test script"`
	if err := os.WriteFile(shScriptPath, []byte(shScriptContent), 0755); err != nil {
		t.Fatal(err)
	}

	pyScriptPath := filepath.Join(tmpDir, "test-script.py")
	pyScriptContent := `print("Hello from test script")`
	if err := os.WriteFile(pyScriptPath, []byte(pyScriptContent), 0755); err != nil {
		t.Fatal(err)
	}

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
	tmpDir := t.TempDir()
	setupTestScripts(t, tmpDir)

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
			sh := createTestShell()

			if strings.Contains(tt.input, "env-script.sh") {
				os.Setenv("TEST_VAR", "test_value")
				defer os.Unsetenv("TEST_VAR")
			}
			oldStdout := os.Stdout
			r, w, _ := os.Pipe() 
			os.Stdout = w

			_, err := sh.executeCommand(tt.input)

			w.Close()
			os.Stdout = oldStdout

			if (err != nil) != tt.wantErr {
				t.Errorf("Shell.executeCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var buf bytes.Buffer
			_, err = buf.ReadFrom(r)
			if err != nil {
				t.Fatalf("Failed to read captured output: %v", err)
			}
			gotOut := buf.String()
			if tt.wantOut != "" {
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
			wantOut: "",
			wantErr: false,
		},
		{
			name:    "cat command with multiple lines",
			input:   "cat",
			stdin:   "line1\nline2\n",
			wantOut: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			stdin := strings.NewReader(tt.stdin)
			sh := createTestShellWithStdin(stdin)

			err := sh.executeSystemCommand(tt.input, []string{}, stdout, stderr)
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

	nonExecPath := filepath.Join(tmpDir, "non-executable.sh")
	err := os.WriteFile(nonExecPath, []byte("#!/bin/sh\necho test\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	sh := createTestShell()
	_, err = sh.executeCommand(nonExecPath)
	if err == nil {
		t.Error("Shell.executeCommand() should fail for non-executable file")
	}
}

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


func TestShell_PwdCommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shell-pwd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe() 
	os.Stdout = w
	sh := createTestShellWithStdin(nil)
	sh.registerCommand(pwd.NewPwdCommand())
	_, err = sh.executeCommand("pwd")
	if err != nil {
		t.Errorf("Shell.executeCommand() error = %v", err)
		return
	}
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Failed to read captured output: %v", err)
	}
	gotOut := strings.Trim(buf.String(), "\n")
	if gotOut != tmpDir {
		t.Errorf("pwd command output = %v, want %v", gotOut, tmpDir)
	}
}

func TestShell_parseCommand(t *testing.T) {
	shell := setupTestShell(t)

	tests := []struct {
		name              string
		input             string
		expectedCmd       string
		expectedArgs      []string
		expectRedirect    bool
		expectedRedirType redirection.RedirectionType
		expectedRedirFile string
		expectedError     error
	}{
		{
			name:              "No input",
			input:             "",
			expectedCmd:       "",
			expectedArgs:      nil,
			expectRedirect:    false,
			expectedRedirType: redirection.NoRedirection,
			expectedError:     nil,
		},
		{
			name:              "Simple command",
			input:             "cmd",
			expectedCmd:       "cmd",
			expectedArgs:      []string{},
			expectRedirect:    false,
			expectedRedirType: redirection.NoRedirection,
			expectedError:     nil,
		},
		{
			name:              "Command with arguments",
			input:             "cmd arg1 arg2",
			expectedCmd:       "cmd",
			expectedArgs:      []string{"arg1", "arg2"},
			expectRedirect:    false,
			expectedRedirType: redirection.NoRedirection,
			expectedError:     nil,
		},
		{
			name:              "Command with double quotes",
			input:             `cmd "quoted arg" arg2`,
			expectedCmd:       "cmd",
			expectedArgs:      []string{`quoted arg`, "arg2"},
			expectRedirect:    false,
			expectedRedirType: redirection.NoRedirection,
			expectedError:     nil,
		},
		{
			name:              "Command with single quotes",
			input:             `cmd 'quoted    arg' arg2`,
			expectedCmd:       "cmd",
			expectedArgs:      []string{`quoted    arg`, "arg2"},
			expectRedirect:    false,
			expectedRedirType: redirection.NoRedirection,
			expectedError:     nil,
		},
		{
			name:              "Output redirection overwrite",
			input:             "cmd > file.txt",
			expectedCmd:       "cmd",
			expectedArgs:      []string{},
			expectRedirect:    true,
			expectedRedirType: redirection.OutputRedirect,
			expectedRedirFile: "file.txt",
			expectedError:     nil,
		},
		{
			name:              "Output redirection append",
			input:             "cmd >> file.txt",
			expectedCmd:       "cmd",
			expectedArgs:      []string{},
			expectRedirect:    true,
			expectedRedirType: redirection.OutputAppend,
			expectedRedirFile: "file.txt",
			expectedError:     nil,
		},
		{
			name:              "Error redirection overwrite",
			input:             "cmd 2> error.log",
			expectedCmd:       "cmd",
			expectedArgs:      []string{},
			expectRedirect:    true,
			expectedRedirType: redirection.ErrorRedirect,
			expectedRedirFile: "error.log",
			expectedError:     nil,
		},
		{
			name:              "Error redirection append",
			input:             "cmd 2>> error.log",
			expectedCmd:       "cmd",
			expectedArgs:      []string{},
			expectRedirect:    true,
			expectedRedirType: redirection.ErrorAppend,
			expectedRedirFile: "error.log",
			expectedError:     nil,
		},
		{
			name:              "Command starting with output redirection",
			input:             "> file3 cat file2", 
			expectedCmd:       "cat",
			expectedArgs:      []string{"file2"},
			expectRedirect:    true,
			expectedRedirType: redirection.OutputRedirect,
			expectedRedirFile: "file3",
			expectedError:     nil,
		},
		{
			name:              "Command starting with output redirection append",
			input:             ">> file3 cat file2",
			expectedCmd:       "cat",
			expectedArgs:      []string{"file2"},
			expectRedirect:    true,
			expectedRedirType: redirection.OutputAppend,
			expectedRedirFile: "file3",
			expectedError:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args, redir, err := shell.parseCommand(tt.input)

			if tt.expectedError != nil {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("parseCommand(%s) error = %v, wantErr %v", tt.input, err, tt.expectedError)
				}
			} else if err != nil {
				t.Fatalf("parseCommand(%s) returned unexpected error: %v", tt.input, err)
			}

			if cmd != tt.expectedCmd {
				t.Errorf("parseCommand(%s) cmd = %q, want %q", tt.input, cmd, tt.expectedCmd)
			}
			if !equalStringSlices(args, tt.expectedArgs) {
				t.Errorf("parseCommand(%s) args = %v, want %v", tt.input, args, tt.expectedArgs)
			}

			if redir == nil {
				t.Fatalf("parseCommand(%s) expected redirection but got nil", tt.input)
			}
			if redir.redirType != tt.expectedRedirType {
				t.Errorf("parseCommand(%s) redirect type = %v, want %v", tt.input, redir.redirType, tt.expectedRedirType)
			}
			if redir.redirType != tt.expectedRedirType {
				t.Errorf("parseCommand(%s) redirect file = %q, want %q", tt.input, redir.redirType, tt.expectedRedirFile)
			}

		})
	}
}

func setupTestShell(t *testing.T) *Shell {
	t.Helper() 
	db := db.GetDB()

	rootDir, err := utils.CurrentPwd()
	if err != nil {
		t.Fatalf("Could not get root directory: %v", err)
	}

	testShell := &Shell{
		user:     user.User{Username: ""}, 
		database: db,
		reader:   bufio.NewReader(&bytes.Buffer{}),
		commands: make(map[string]command.Command),
		history:  make(map[string]int),
		rootDir:  rootDir,
	}

	exitCmd := exit.NewExitCommand(testShell.database, &testShell.user)
	testShell.registerCommand(exitCmd)
	echoCmd := echo.NewEchoCommand()
	testShell.registerCommand(echoCmd)
	catCmd := cat.NewCatCommand()
	testShell.registerCommand(catCmd)
	pwdCmd := pwd.NewPwdCommand()
	testShell.registerCommand(pwdCmd)
	cdCmd := cd.NewCDCommand(testShell.rootDir)
	testShell.registerCommand(cdCmd)
	lsCmd := ls.NewLSCommand()
	testShell.commands[lsCmd.Name()] = lsCmd
	colorCmd := color.NewColorCommand()
	testShell.commands[colorCmd.Name()] = colorCmd
	loginCmd := login.NewLoginCommand(testShell.database, &testShell.user)
	testShell.commands[loginCmd.Name()] = loginCmd
	adduserCmd := adduser.NewAddUserCommand(testShell.database, &testShell.user)
	testShell.commands[adduserCmd.Name()] = adduserCmd
	logoutCmd := logout.NewLogoutCommand(testShell.database, &testShell.user)
	testShell.commands[logoutCmd.Name()] = logoutCmd
	historyCmd := history.NewHistoryCommand(&testShell.history, &testShell.user, testShell.database)
	testShell.commands[historyCmd.Name()] = historyCmd
	helpCmd := help.NewHelpCommand()
	testShell.commands[helpCmd.Name()] = helpCmd
	shellBuiltins := []string{}
	for cmd := range testShell.commands {
		shellBuiltins = append(shellBuiltins, cmd)
	}
	typeCmd := typecmd.NewTypeCommand(shellBuiltins)
	testShell.registerCommand(typeCmd)

	stdoutBuf := &bytes.Buffer{}
	testShell.commands["pwd"].Execute([]string{}, stdoutBuf)
	testShell.rootDir = strings.TrimSpace(stdoutBuf.String()) 

	return testShell
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func mockShell(input string) (*Shell, *bytes.Buffer) {
	inputReader := strings.NewReader(input)
	outputBuffer := new(bytes.Buffer)

	shell := New()
	shell.reader = bufio.NewReader(inputReader)

	return shell, outputBuffer
}

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

func createTestShell() *Shell {
	sh := New()
	return sh
}
func createTestShellWithStdin(stdin io.Reader) *Shell {
	sh := New()
	if stdin != nil {
		sh.reader = bufio.NewReader(stdin) 
	}
	return sh
}
