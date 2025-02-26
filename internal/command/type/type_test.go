package typecmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"asa/shell/utils"
)

func TestTypeCommand_Execute(t *testing.T) {
	tmpDir := t.TempDir()

	executableName := "test_executable"
	if runtime.GOOS == "windows" {
		executableName += ".exe"
	}
	executablePath := filepath.Join(tmpDir, executableName)
	err := os.WriteFile(executablePath, []byte("#!/bin/bash\necho test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create dummy executable: %v", err)
	}

	testCases := []struct {
		name           string
		input          []string
		builtins       []string
		pathEnv        string
		expectedOutput string
		wantErr        error
		setupEnv       func() func() 
	}{
		{
			name:           "No arguments",
			input:          []string{},
			expectedOutput: "",
			wantErr:        utils.ErrMissingCommandName,
		},
		{
			name:           "Built-in command",
			input:          []string{"echo"},
			builtins:       []string{"echo", "pwd"},
			expectedOutput: "echo is a shell builtin\n",
			wantErr:        nil,
		},
		{
			name:           "External command in PATH",
			input:          []string{"test_executable"},
			builtins:       []string{"echo"},
			pathEnv:        tmpDir,
			expectedOutput: fmt.Sprintf("test_executable is %s\n", executablePath),
			wantErr:        nil,
			setupEnv: func() func() {
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+originalPath)
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
		},
		{
			name:           "External command not in PATH",
			input:          []string{"non_existent_command"},
			builtins:       []string{"echo"},
			pathEnv:        tmpDir,
			expectedOutput: "",
			wantErr:        utils.ErrCommandNotFound,
		},
		{
			name:           "PATH not set, external command",
			input:          []string{"ls"},
			builtins:       []string{"echo"},
			pathEnv:        "",
			expectedOutput: "",
			wantErr:        utils.ErrEnvironmentVarNotSet,
			setupEnv: func() func() {
				originalPath := os.Getenv("PATH")
				os.Unsetenv("PATH")
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
		},
		{
			name:           "Multiple arguments - mixed types",
			input:          []string{"echo", "test_executable", "non_existent_command"},
			builtins:       []string{"echo", "pwd"},
			pathEnv:        tmpDir,
			expectedOutput: "echo is a shell builtin\n" + fmt.Sprintf("test_executable is %s\n", executablePath),
			wantErr:        utils.ErrCommandNotFound, 
			setupEnv: func() func() {
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+originalPath)
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
		},
		{
			name:           "External command in different case (PATH sensitive)",
			input:          []string{"Test_Executable"},
			builtins:       []string{"echo"},
			pathEnv:        tmpDir,
			expectedOutput: "",
			wantErr:        utils.ErrCommandNotFound,
			setupEnv: func() func() {
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+originalPath)
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupEnv != nil {
				teardownEnv := tc.setupEnv()
				defer teardownEnv()
			}
			if tc.pathEnv != "" {
				os.Setenv("PATH", tc.pathEnv)
			} else if tc.setupEnv == nil && tc.pathEnv == "" && strings.Contains(tc.name, "PATH not set") {
				os.Unsetenv("PATH")
			}

			cmd := NewTypeCommand(tc.builtins)
			var outBuf bytes.Buffer
			err := cmd.Execute(tc.input, &outBuf)

			if tc.wantErr != nil {
				if err == nil {
					t.Errorf("Test case '%s': Expected error '%v', but got nil", tc.name, tc.wantErr)
				} else if !errors.Is(err, tc.wantErr) {
					t.Errorf("Test case '%s': Expected error type '%v', but got '%v'", tc.name, tc.wantErr, err)
				}
			} else if err != nil {
				t.Errorf("Test case '%s': Unexpected error: %v", tc.name, err)
			}

			actualOutput := outBuf.String()
			if actualOutput != tc.expectedOutput {
				t.Errorf("Test case '%s': Output mismatch:\nexpected:\n'%s'\ngot:\n'%s'", tc.name, tc.expectedOutput, actualOutput)
			}

			if tc.setupEnv == nil && tc.pathEnv != "" {
				os.Unsetenv("PATH")
			} else if tc.setupEnv == nil && tc.pathEnv == "" && strings.Contains(tc.name, "PATH not set") {
			}
		})
	}
}

func TestTypeCommand_findCommand(t *testing.T) {
	tmpDir := t.TempDir()

	executableName := "test_executable"
	if runtime.GOOS == "windows" {
		executableName += ".exe"
	}
	executablePath := filepath.Join(tmpDir, executableName)
	err := os.WriteFile(executablePath, []byte("#!/bin/bash\necho test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create dummy executable: %v", err)
	}

	testCases := []struct {
		name           string
		cmdInput       string
		builtins       []string
		pathEnv        string
		expectedOutput string
		wantErr        error
		setupEnv       func() func()
	}{
		{
			name:           "Builtin command - findCommand",
			cmdInput:       "echo",
			builtins:       []string{"echo"},
			expectedOutput: "", 
			wantErr:        nil,
		},
		{
			name:           "External command in PATH - findCommand",
			cmdInput:       "test_executable",
			builtins:       []string{},
			pathEnv:        tmpDir,
			expectedOutput: fmt.Sprintf("test_executable is %s", executablePath),
			wantErr:        nil,
			setupEnv: func() func() {
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+originalPath)
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
		},
		{
			name:           "External command not in PATH - findCommand",
			cmdInput:       "non_existent_command",
			builtins:       []string{},
			pathEnv:        tmpDir,
			expectedOutput: "",
			wantErr:        utils.ErrCommandNotFound,
		},
		{
			name:           "PATH not set - findCommand",
			cmdInput:       "ls",
			builtins:       []string{},
			pathEnv:        "",
			expectedOutput: "",
			wantErr:        utils.ErrEnvironmentVarNotSet,
			setupEnv: func() func() {
				originalPath := os.Getenv("PATH")
				os.Unsetenv("PATH")
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupEnv != nil {
				teardownEnv := tc.setupEnv()
				defer teardownEnv()
			}
			if tc.pathEnv != "" {
				os.Setenv("PATH", tc.pathEnv)
			} else if tc.setupEnv == nil && tc.pathEnv == "" && strings.Contains(tc.name, "PATH not set") {
				os.Unsetenv("PATH") 
			}

			cmd := NewTypeCommand(tc.builtins)

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() {
				os.Stdout = oldStdout
			}()

			outputChan := make(chan string)
			go func() {
				var buf bytes.Buffer
				io.Copy(&buf, r)
				outputChan <- buf.String()
			}()

			resultOutput, err := cmd.findCommand(tc.cmdInput)

			w.Close()
			capturedOutput := <-outputChan

			if tc.wantErr != nil {
				if err == nil {
					t.Errorf("Test case '%s': Expected error '%v', but got nil", tc.name, tc.wantErr)
				} else if !errors.Is(err, tc.wantErr) {
					t.Errorf("Test case '%s': Expected error type '%v', but got '%v'", tc.name, tc.wantErr, err)
				}
			} else if err != nil {
				t.Errorf("Test case '%s': Unexpected error: %v", tc.name, err)
			}

			expectedFullOutput := tc.expectedOutput
			if tc.builtins != nil && containsString(tc.builtins, tc.cmdInput) {
				expectedFullOutput = fmt.Sprintf("%s is a shell builtin", tc.cmdInput) 
			}

			fullOutput := capturedOutput + resultOutput 

			if !strings.Contains(fullOutput, expectedFullOutput) && expectedFullOutput != "" { 
				t.Errorf("Test case '%s': Output mismatch:\nexpected to contain:\n'%s'\ngot:\n'%s'", tc.name, expectedFullOutput, fullOutput)
			} else if expectedFullOutput == "" && fullOutput != "" { 
				t.Errorf("Test case '%s': Output mismatch:\nexpected empty string, but got:\n'%s'", tc.name, fullOutput)
			}

			if tc.setupEnv == nil && tc.pathEnv != "" {
				os.Unsetenv("PATH")
			} else if tc.setupEnv == nil && tc.pathEnv == "" && strings.Contains(tc.name, "PATH not set") {
			}
		})
	}
}

func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
