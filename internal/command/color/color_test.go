package color

import (
	"asa/shell/utils"
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestColorCommand_Execute(t *testing.T) {

	tests := []struct {
		name               string
		inputArgs          []string
		initialColorState  bool
		expectedColorState bool
		expectedStdout     string
		expectedError      bool
	}{
		{
			name:               "Color On",
			inputArgs:          []string{"on"},
			initialColorState:  false,
			expectedColorState: true,
			expectedStdout:     utils.ColorText("Color is set on", utils.TextBlue),
			expectedError:      false,
		},
		{
			name:               "Color Off",
			inputArgs:          []string{"off"},
			initialColorState:  true,
			expectedColorState: false,
			expectedStdout:     "Color is set off",
			expectedError:      false,
		},
		{
			name:               "Color On - Already On",
			inputArgs:          []string{"on"},
			initialColorState:  true,
			expectedColorState: true,
			expectedStdout:     "color is already set\n",
			expectedError:      true,
		},
		{
			name:               "Color Off - Already Off",
			inputArgs:          []string{"off"},
			initialColorState:  false,
			expectedColorState: false,
			expectedStdout:     "color is not set\n",
			expectedError:      true,
		},
		{
			name:               "Color Invalid Argument",
			inputArgs:          []string{"invalid"},
			initialColorState:  false,
			expectedColorState: false, // Invalid arg should not change state
			expectedStdout:     "",
			expectedError:      true,
		},
		{
			name:               "Color No Argument",
			inputArgs:          []string{},
			initialColorState:  false,
			expectedColorState: false,
			expectedStdout:     "",
			expectedError:      true, // No error for status check with no args
		},
		{
			name:               "Color Case Insensitive On",
			inputArgs:          []string{"ON"},
			initialColorState:  false,
			expectedColorState: true,
			expectedStdout:     utils.ColorText("Color is set on", utils.TextBlue),
			expectedError:      false,
		},
		{
			name:               "Color Case Insensitive Off",
			inputArgs:          []string{"Off"},
			initialColorState:  true,
			expectedColorState: false,
			expectedStdout:     "Color is set off",
			expectedError:      false,
		},
		{
			name:               "Color Extra Arguments",
			inputArgs:          []string{"on", "extra"},
			initialColorState:  false,
			expectedColorState: false, // Extra args should not change state and return error
			expectedStdout:     "",
			expectedError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewColorCommand()

			if got := cmd.Name(); got != "color" {
				t.Errorf("CatCommand.Name() = %v, want %v", got, "cat")
			}

			if tt.initialColorState {
				os.Setenv("SHELLCOLOR", "1")
			} else {
				if _, ok := os.LookupEnv("SHELLCOLOR"); ok {
					os.Unsetenv("SHELLCOLOR")
				}
			}
			stdout := &bytes.Buffer{}
			err := cmd.Execute(tt.inputArgs, stdout)

			if (err != nil) && tt.expectedError {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check output
			gotOutput := strings.TrimSpace(stdout.String())
			expectedOutput := strings.TrimSpace(tt.expectedStdout)
			if !tt.expectedError {
				if utils.IsColor() != tt.expectedColorState {
					t.Errorf("Execute() color state = %v, want %v", utils.IsColor(), tt.expectedColorState)
				}
				if gotOutput != expectedOutput {
					t.Errorf("Execute() stdout = %q, want %q", gotOutput, expectedOutput)
				}
			}
		})
	}
}
