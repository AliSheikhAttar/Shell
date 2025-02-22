package command

import (
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
		{
			name:    "valid argument 1",
			args:    []string{"1"},
			wantErr: nil,
		},
        {
			name:    "valid argument 2",
			args:    []string{"2"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewExitCommand()

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
