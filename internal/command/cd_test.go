package command

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCDCommand(t *testing.T) {
	// Save current directory to restore it after tests
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Get user's home directory for testing
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		args    []string
		wantDir string
		wantErr error
		setup   func() error
	}{
		{
			name:    "no arguments - go to home directory",
			args:    []string{},
			wantDir: homeDir,
		},
		{
			name:    "tilde - go to home directory",
			args:    []string{"~"},
			wantDir: homeDir,
		},
		{
			name:    "absolute path",
			args:    []string{tempDir},
			wantDir: tempDir,
		},
		{
			name:    "relative path",
			args:    []string{".."},
			wantDir: filepath.Dir(originalDir),
		},
		{
			name:    "path with tilde",
			args:    []string{"~/Documents"},
			wantDir: filepath.Join(homeDir, "Documents"),
			setup: func() error {
				// Create Documents directory if it doesn't exist
				return os.MkdirAll(filepath.Join(homeDir, "Documents"), 0755)
			},
		},
		{
			name:    "too many arguments",
			args:    []string{"dir1", "dir2"},
			wantErr: ErrTooManyArgs,
		},
		{
			name:    "non-existent directory",
			args:    []string{filepath.Join(tempDir, "nonexistent")},
			wantErr: &os.PathError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to original directory before each test
			if err := os.Chdir(originalDir); err != nil {
				t.Fatalf("Failed to reset directory: %v", err)
			}

			// Run setup if provided
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			cmd := NewCDCommand()

			// Verify command name
			if got := cmd.Name(); got != "cd" {
				t.Errorf("CDCommand.Name() = %v, want %v", got, "cd")
			}

			// Execute command
			err := cmd.Execute(tt.args, os.Stdout)

			// Check error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("CDCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expect an error of a specific type, check it
			if tt.wantErr != nil {
				if _, ok := err.(interface{ Error() string }); !ok {
					t.Errorf("Expected error of type %T, got %T", tt.wantErr, err)
				}
				return
			}

			// If no error, verify current directory
			if err == nil {
				got, err := os.Getwd()
				if err != nil {
					t.Fatalf("Failed to get current directory: %v", err)
				}
				if got != tt.wantDir {
					t.Errorf("Current directory = %v, want %v", got, tt.wantDir)
				}
			}
		})
	}
}
