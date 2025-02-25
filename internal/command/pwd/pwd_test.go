package pwd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPwdCommand_Execute(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "pwd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original directory and PWD
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	originalPWD := os.Getenv("PWD")
	defer func() {
		os.Chdir(originalDir)
		if originalPWD != "" {
			os.Setenv("PWD", originalPWD)
		}
	}()

	tests := []struct {
		name      string
		setupFunc func() string
		args      []string
		wantErr   bool
	}{
		{
			name: "using PWD env var",
			setupFunc: func() string {
				os.Chdir(tmpDir)
				os.Setenv("PWD", tmpDir)
				return tmpDir
			},
			args:    []string{},
			wantErr: false,
		},
		{
			name: "without PWD env var",
			setupFunc: func() string {
				os.Chdir(tmpDir)
				os.Unsetenv("PWD")
				return tmpDir
			},
			args:    []string{},
			wantErr: false,
		},
		{
			name: "with symlink directory",
			setupFunc: func() string {
				symDir := filepath.Join(tmpDir, "symlink")
				os.Mkdir(filepath.Join(tmpDir, "real"), 0755)
				os.Symlink(filepath.Join(tmpDir, "real"), symDir)
				os.Chdir(symDir)
				return filepath.Join(tmpDir, "real")
			},
			args:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedDir := tt.setupFunc()

			stdout := &bytes.Buffer{}
			cmd := NewPwdCommand()

			err := cmd.Execute(tt.args, stdout)
			if (err != nil) != tt.wantErr {
				t.Errorf("PwdCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := strings.TrimSpace(stdout.String())
			if got != expectedDir {
				t.Errorf("PwdCommand.Execute() output = %v, want %v", got, expectedDir)
			}
		})
	}
}
