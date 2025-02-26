package cd

import (
	"asa/shell/utils"
	"os"
	"path/filepath"
	"testing"
)

func TestCDCommand(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

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
			wantErr: utils.ErrTooManyArgs,
		},
		{
			name:    "non-existent directory",
			args:    []string{filepath.Join(tempDir, "nonexistent")},
			wantErr: &os.PathError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Chdir(originalDir); err != nil {
				t.Fatalf("Failed to reset directory: %v", err)
			}

			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			cmd := NewCDCommand(homeDir)

			// Verify command name
			if got := cmd.Name(); got != "cd" {
				t.Errorf("CDCommand.Name() = %v, want %v", got, "cd")
			}

			err := cmd.Execute(tt.args, os.Stdout)

			// Check error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("CDCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				if _, ok := err.(interface{ Error() string }); !ok {
					t.Errorf("Expected error of type %T, got %T", tt.wantErr, err)
				}
				return
			}

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
