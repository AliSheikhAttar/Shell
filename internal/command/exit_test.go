package command

import (
	"os"
	"testing"
)

func TestExitCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr error
	}{
		{
			name:    "too many arguments",
			args:    []string{"1", "2"},
			wantErr: ErrTooManyArgs,
		},
		{
			name:    "invalid argument",
			args:    []string{"invalid"},
			wantErr: ErrInvalidArgs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewExitCommand(os.Stdout)

			// Verify command name
			if got := cmd.Name(); got != "exit" {
				t.Errorf("ExitCommand.Name() = %v, want %v", got, "exit")
			}

			// Test error cases
			err := cmd.Execute(tt.args)
			if err != tt.wantErr {
				t.Errorf("ExitCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
