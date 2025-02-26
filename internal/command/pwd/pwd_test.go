package pwd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPwdCommand_Execute(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
		wantErr        error
		setup          func() (string, string, func()) 
	}{
		{
			name:           "Basic execution",
			args:           []string{},
			expectedOutput: "",
			wantErr:        nil,
			setup: func() (string, string, func()) {
				initialDir, _ := os.Getwd()
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				expectedOutput := tempDir + "\n"
				return initialDir, expectedOutput, func() { 
					os.Chdir(initialDir)
				}
			},
		},
		{
			name:           "No arguments",
			args:           []string{"extra_arg"}, 
			expectedOutput: "",                    
			wantErr:        nil,
			setup: func() (string, string, func()) {
				initialDir, _ := os.Getwd()
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				expectedOutput := tempDir + "\n"
				return initialDir, expectedOutput, func() { 
					os.Chdir(initialDir)
				}
			},
		},

	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initialDir, expectedOutput, teardown := tc.setup() 
			defer teardown()

			var outBuf bytes.Buffer
			cmd := NewPwdCommand()
			err := cmd.Execute(tc.args, &outBuf)

			if tc.wantErr != nil {
				if err == nil {
					t.Errorf("Test case '%s': Expected error '%v', but got nil", tc.name, tc.wantErr)
				} else if !errors.Is(err, tc.wantErr) {
					t.Errorf("Test case '%s': Expected error type '%v', but got '%v'", tc.name, tc.wantErr, err)
				}
			} else if err != nil {
				t.Errorf("Test case '%s': Unexpected error: %v", tc.name, err)
			}

			if expectedOutput == "" {
				currentDir, _ := os.Getwd() 
				expectedOutput = currentDir + "\n"
			}

			actualOutput := outBuf.String()
			if actualOutput != expectedOutput {
				t.Errorf("Test case '%s': Output mismatch:\nexpected:\n'%s'\ngot:\n'%s'", tc.name, expectedOutput, actualOutput)
			}

			os.Chdir(initialDir)
			currentDirAfterTest, _ := os.Getwd()
			if currentDirAfterTest != initialDir {
				t.Errorf("Test case '%s': Initial directory was not restored. Started in '%s', ended in '%s'", tc.name, initialDir, currentDirAfterTest)
			}

		})
	}
}

func TestPwdCommand_GetCurrentDirectory(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func() (string, string, func()) 
		expectedOutput string
		wantErr        bool
	}{
		{
			name: "Using filepath.Abs",
			setup: func() (string, string, func()) {
				initialDir, _ := os.Getwd()
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				expectedOutput := tempDir 
				return initialDir, expectedOutput, func() {
					os.Chdir(initialDir)
				}
			},
			expectedOutput: "", 
			wantErr:        false,
		},

		{
			name: "Invalid PWD env variable, fallback to Abs",
			setup: func() (string, string, func()) {
				initialDir, _ := os.Getwd()
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				os.Setenv("PWD", "/invalid/directory") 
				expectedOutput := tempDir              
				return initialDir, expectedOutput, func() {
					os.Chdir(initialDir)
					os.Unsetenv("PWD")
				}
			},
			expectedOutput: "", 
			wantErr:        false,
		},
		{
			name: "Valid PWD env variable, valid directory",
			setup: func() (string, string, func()) {
				initialDir, _ := os.Getwd()
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				os.Setenv("PWD", tempDir) 
				expectedOutput := tempDir 
				return initialDir, expectedOutput, func() {
					os.Chdir(initialDir)
					os.Unsetenv("PWD")
				}
			},
			expectedOutput: "", 
			wantErr:        false,
		},
	}

	if runtime.GOOS == "linux" { 
		testCases = append(testCases, struct {
			name           string
			setup          func() (string, string, func()) 
			expectedOutput string
			wantErr        bool
		}{
			name: "Using /proc/self/cwd (Linux)",
			setup: func() (string, string, func()) {
				initialDir, _ := os.Getwd()
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				expectedOutput := tempDir 
				return initialDir, expectedOutput, func() {
					os.Chdir(initialDir)
				}
			},
			expectedOutput: "", 
			wantErr:        false,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initialDir, expectedOutput, teardown := tc.setup()
			defer teardown()

			cmd := NewPwdCommand()
			pwd, err := cmd.getCurrentDirectory()

			if tc.wantErr {
				if err == nil {
					t.Errorf("Test case '%s': Expected error, but got nil", tc.name)
				}
			} else if err != nil {
				t.Errorf("Test case '%s': Unexpected error: %v", tc.name, err)
			}

			if expectedOutput == "" {
				currentDir, _ := os.Getwd() 
				expectedOutput = currentDir
			}

			if pwd != expectedOutput {
				t.Errorf("Test case '%s': Output mismatch:\nexpected:\n'%s'\ngot:\n'%s'", tc.name, expectedOutput, pwd)
			}
			os.Chdir(initialDir)
			currentDirAfterTest, _ := os.Getwd()
			if currentDirAfterTest != initialDir {
				t.Errorf("Test case '%s': Initial directory was not restored. Started in '%s', ended in '%s'", tc.name, initialDir, currentDirAfterTest)
			}
		})
	}
}

func TestPwdCommand_GetCurrentDirectory_PathCleaning(t *testing.T) {
	initialDir, _ := os.Getwd()
	defer os.Chdir(initialDir)

	tempDir := t.TempDir()
	os.Chdir(tempDir)

	// Create a path with extra separators and ".." elements
	messyPath := filepath.Join(tempDir, "dir1", ".", "dir2", "..", "dir3")
	os.MkdirAll(messyPath, 0755)
	os.Chdir(messyPath)

	cmd := NewPwdCommand()
	pwd, err := cmd.getCurrentDirectory()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	cleanedExpectedPath := tempDir 

	if pwd != messyPath {
		t.Errorf("Path cleaning test failed: Expected cleaned path to be '%s', but got '%s'", cleanedExpectedPath, pwd)
	}
	os.Chdir(initialDir)
	currentDirAfterTest, _ := os.Getwd()
	if currentDirAfterTest != initialDir {
		t.Errorf("Initial directory was not restored. Started in '%s', ended in '%s'", initialDir, currentDirAfterTest)
	}
}

func TestPwdCommand_Name(t *testing.T) {
	cmd := NewPwdCommand()
	if cmd.Name() != "pwd" {
		t.Errorf("Name() should return 'pwd', but got '%s'", cmd.Name())
	}
}
