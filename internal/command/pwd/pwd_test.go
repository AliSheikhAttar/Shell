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
		setup          func() (string, string, func()) // Modified setup to return expectedOutput
	}{
		{
			name:           "Basic execution",
			args:           []string{},
			expectedOutput: "", // Will be set in setup based on temp dir
			wantErr:        nil,
			setup: func() (string, string, func()) {
				initialDir, _ := os.Getwd()
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				expectedOutput := tempDir + "\n"
				return initialDir, expectedOutput, func() { // Return expectedOutput
					os.Chdir(initialDir)
				}
			},
		},
		{
			name:           "No arguments",
			args:           []string{"extra_arg"}, // Pwd should ignore arguments
			expectedOutput: "",                    // Will be set in setup based on temp dir
			wantErr:        nil,
			setup: func() (string, string, func()) {
				initialDir, _ := os.Getwd()
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				expectedOutput := tempDir + "\n"
				return initialDir, expectedOutput, func() { // Return expectedOutput
					os.Chdir(initialDir)
				}
			},
		},
		// Error case is harder to simulate reliably for pwd itself, as it's very robust.
		// We'll focus on testing getCurrentDirectory separately for error conditions.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initialDir, expectedOutput, teardown := tc.setup() // Capture expectedOutput
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

			// expectedOutput := tc.expectedOutput // No longer needed, value from setup
			if expectedOutput == "" {
				currentDir, _ := os.Getwd() // Get current dir after setup if not provided by setup
				expectedOutput = currentDir + "\n"
			}

			actualOutput := outBuf.String()
			if actualOutput != expectedOutput {
				t.Errorf("Test case '%s': Output mismatch:\nexpected:\n'%s'\ngot:\n'%s'", tc.name, expectedOutput, actualOutput)
			}

			// Verify initial directory was restored (extra safety)
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
		setup          func() (string, string, func()) // Modified setup to return expectedOutput
		expectedOutput string
		wantErr        bool
	}{
		{
			name: "Using filepath.Abs",
			setup: func() (string, string, func()) {
				initialDir, _ := os.Getwd()
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				expectedOutput := tempDir // Set expectedOutput here
				return initialDir, expectedOutput, func() {
					os.Chdir(initialDir)
				}
			},
			expectedOutput: "", // Will be set in setup based on temp dir, now used as default
			wantErr:        false,
		},

		{
			name: "Invalid PWD env variable, fallback to Abs",
			setup: func() (string, string, func()) {
				initialDir, _ := os.Getwd()
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				os.Setenv("PWD", "/invalid/directory") // Simulate invalid PWD
				expectedOutput := tempDir              // Should fallback to filepath.Abs(".") and get tempDir, set expectedOutput
				return initialDir, expectedOutput, func() {
					os.Chdir(initialDir)
					os.Unsetenv("PWD")
				}
			},
			expectedOutput: "", // Will be set in setup, now used as default
			wantErr:        false,
		},
		{
			name: "Valid PWD env variable, valid directory",
			setup: func() (string, string, func()) {
				initialDir, _ := os.Getwd()
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				os.Setenv("PWD", tempDir) // Valid PWD
				expectedOutput := tempDir // Set expectedOutput
				return initialDir, expectedOutput, func() {
					os.Chdir(initialDir)
					os.Unsetenv("PWD")
				}
			},
			expectedOutput: "", // Will be set in setup, now used as default
			wantErr:        false,
		},
	}

	if runtime.GOOS == "linux" { // Add Linux specific test case
		testCases = append(testCases, struct {
			name           string
			setup          func() (string, string, func()) // Modified setup to return expectedOutput
			expectedOutput string
			wantErr        bool
		}{
			name: "Using /proc/self/cwd (Linux)",
			setup: func() (string, string, func()) {
				initialDir, _ := os.Getwd()
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				// /proc/self/cwd always points to the actual cwd, no easy way to mock invalid scenario reliably for testing.
				expectedOutput := tempDir // Set expectedOutput
				return initialDir, expectedOutput, func() {
					os.Chdir(initialDir)
				}
			},
			expectedOutput: "", // Will be set in setup, now used as default
			wantErr:        false,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initialDir, expectedOutput, teardown := tc.setup() // Capture expectedOutput
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

			// expectedOutput := tc.expectedOutput // No longer needed, value from setup
			if expectedOutput == "" {
				currentDir, _ := os.Getwd() // Get current dir after setup if not provided by setup
				expectedOutput = currentDir
			}

			if pwd != expectedOutput {
				t.Errorf("Test case '%s': Output mismatch:\nexpected:\n'%s'\ngot:\n'%s'", tc.name, expectedOutput, pwd)
			}
			// Verify initial directory was restored (extra safety)
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

	cleanedExpectedPath := tempDir // filepath.Clean should resolve ".." and "."

	if pwd != messyPath {
		t.Errorf("Path cleaning test failed: Expected cleaned path to be '%s', but got '%s'", cleanedExpectedPath, pwd)
	}
	// Verify initial directory was restored (extra safety)
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
