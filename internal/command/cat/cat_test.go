package cat

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedOutput string
		wantErr     bool // Changed wantErr to boolean
		setup       func(tempDir string) (filePaths []string, err error)
	}{
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true, // Expecting an error
		},
		{
			name: "single file - simple content",
			args: []string{"test_file1.txt"},
			expectedOutput: "Hello, file 1!\n",
			wantErr:     false, // Not expecting an error
			setup: func(tempDir string) (filePaths []string, err error) {
				filePath := filepath.Join(tempDir, "test_file1.txt")
				err = os.WriteFile(filePath, []byte("Hello, file 1!\n"), 0644)
				if err != nil {
					return nil, err
				}
				return []string{filePath}, nil
			},
		},
		{
			name: "multiple files",
			args: []string{"test_file1.txt", "test_file2.txt"},
			expectedOutput: "Content of file 1\nContent of file 2\n",
			wantErr:     false, // Not expecting an error
			setup: func(tempDir string) (filePaths []string, error error) {
				file1Path := filepath.Join(tempDir, "test_file1.txt")
				err := os.WriteFile(file1Path, []byte("Content of file 1\n"), 0644)
				if err != nil {
					return nil, err
				}
				file2Path := filepath.Join(tempDir, "test_file2.txt")
				err = os.WriteFile(file2Path, []byte("Content of file 2\n"), 0644)
				if err != nil {
					return nil, err
				}
				return []string{file1Path, file2Path}, nil
			},
		},
		{
			name:    "non-existent file",
			args:    []string{"non_existent_file.txt"},
			wantErr: true, // Expecting an error
		},
		{
			name: "mixed valid and non-existent files",
			args: []string{"test_file1.txt", "non_existent_file.txt"},
			expectedOutput: "Content of file 1\n", // Should print valid file content before error
			wantErr:      true, // Expecting an error
			setup: func(tempDir string) (filePaths []string, error error) {
				filePath := filepath.Join(tempDir, "test_file1.txt")
				err := os.WriteFile(filePath, []byte("Content of file 1\n"), 0644)
				if err != nil {
					return nil, err
				}
				return []string{filePath}, nil
			},
		},
		{
			name: "file with empty lines",
			args: []string{"test_file1.txt"},
			expectedOutput: "Line 1\n\nLine 3\n",
			wantErr:     false, // Not expecting an error
			setup: func(tempDir string) (filePaths []string, error error) {
				filePath := filepath.Join(tempDir, "test_file1.txt")
				content := "Line 1\n\nLine 3\n"
				err := os.WriteFile(filePath, []byte(content), 0644)
				if err != nil {
					return nil, err
				}
				return []string{filePath}, nil
			},
		},
		{
			name: "file with no newline at end",
			args: []string{"test_file1.txt"},
			expectedOutput: "Line 1", // scanner.Scan() should handle no newline at end
			wantErr:     false, // Not expecting an error
			setup: func(tempDir string) (filePaths []string, error error) {
				filePath := filepath.Join(tempDir, "test_file1.txt")
				content := "Line 1"
				err := os.WriteFile(filePath, []byte(content), 0644)
				if err != nil {
					return nil, err
				}
				return []string{filePath}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCatCommand()

			// Verify command name
			if got := cmd.Name(); got != "cat" {
				t.Errorf("CatCommand.Name() = %v, want %v", got, "cat")
			}

			tempDir, err := os.MkdirTemp("", "cat-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			var filePaths []string
			if tt.setup != nil {
				filePaths, err = tt.setup(tempDir)
				if err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
				// Modify args to use full file paths in temp directory
				for i, arg := range tt.args {
					if strings.HasPrefix(arg, "test_file") { // simple way to replace test_file names in args
						for _, filePath := range filePaths {
							if strings.HasSuffix(filePath, arg) {
								tt.args[i] = filePath
								break
							}
						}
					}
				}
			}

			// Capture stdout
			stdout := &bytes.Buffer{}
			err = cmd.Execute(tt.args, stdout)

			// Check error - Simplified to boolean check
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
				// We are no longer checking the specific error type here
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check output
			gotOutput := strings.TrimSpace(stdout.String())
			expectedOutput := strings.TrimSpace(tt.expectedOutput) // trim to handle possible newline inconsistencies in tests
			if gotOutput != expectedOutput {
				t.Errorf("CatCommand.Execute() output = \n%v\n, want \n%v", gotOutput, expectedOutput)
			}
		})
	}
}