package help

import (
	"bytes"
	"errors"
	"testing"

	"asa/shell/utils"
)

func TestHelpCommand_Name(t *testing.T) {
	cmd := NewHelpCommand()
	if cmd.Name() != "help" {
		t.Errorf("Name() should return '--help', but got '%s'", cmd.Name())
	}
}

func TestHelpCommand_Execute(t *testing.T) {
	testCases := []struct {
		name    string
		args    []string
		wantErr error
	}{
		{
			name:    "No arguments",
			args:    []string{},
			wantErr: nil,
		},
		{
			name:    "Too many arguments",
			args:    []string{"extra"},
			wantErr: utils.ErrInvalidArgs,
		},
		{
			name:    "Multiple extra arguments",
			args:    []string{"arg1", "arg2"},
			wantErr: utils.ErrInvalidArgs,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewHelpCommand()
			var outBuf bytes.Buffer
			err := cmd.Execute(tc.args, &outBuf)

			if tc.wantErr != nil {
				if err == nil {
					t.Errorf("Test case '%s': Expected error '%v', but got nil", tc.name, tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("Test case '%s': Expected error type '%v', but got '%v'", tc.name, tc.wantErr, err)
				}
				return // Skip output check if error is expected
			}

			if err != nil {
				t.Fatalf("Test case '%s': Unexpected error: %v", tc.name, err)
			}

		})
	}
}
