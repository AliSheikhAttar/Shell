package color

import (
	"asa/shell/utils"
	"bytes"
	"errors"
	"os"
	"testing"
)

var ErrInvalidArgument = errors.New("invalid argument")

func TestColorText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		color    string
		expected string
	}{
		{
			name:     "red text",
			text:     "error",
			color:    utils.TextRed,
			expected: utils.TextRed + "error" + utils.Reset,
		},
		{
			name:     "green text",
			text:     "success",
			color:    utils.TextGreen,
			expected: utils.TextGreen + "success" + utils.Reset,
		},
		{
			name:     "blue text with background",
			text:     "info",
			color:    utils.TextBlue + utils.BgWhite,
			expected: utils.TextBlue + utils.BgWhite + "info" + utils.Reset,
		},
		{
			name:     "formatted text",
			text:     "important",
			color:    utils.Bold + utils.TextYellow,
			expected: utils.Bold + utils.TextYellow + "important" + utils.Reset,
		},
		{
			name:     "empty text",
			text:     "",
			color:    utils.TextCyan,
			expected: utils.TextCyan + "" + utils.Reset,
		},
		{
			name:     "multiple formatting",
			text:     "warning",
			color:    utils.Bold + utils.Underline + utils.TextMagenta,
			expected: utils.Bold + utils.Underline + utils.TextMagenta + "warning" + utils.Reset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ColorText(tt.text, tt.color)
			if result != tt.expected {
				t.Errorf("utils.ColorText() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsColor(t *testing.T) {
	tests := []struct {
		name        string
		envSet      bool
		envValue    string
		wantEnabled bool
	}{
		{
			name:        "SHELLCOLOR environment variable set",
			envSet:      true,
			envValue:    "1",
			wantEnabled: true,
		},
		{
			name:        "SHELLCOLOR environment variable not set",
			envSet:      false,
			envValue:    "",
			wantEnabled: false,
		},
		{
			name:        "SHELLCOLOR environment variable set to empty",
			envSet:      true,
			envValue:    "",
			wantEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variable before test
			os.Unsetenv("SHELLCOLOR")

			// Set environment variable if test requires it
			if tt.envSet {
				os.Setenv("SHELLCOLOR", tt.envValue)
			}

			// Run test
			got := utils.IsColor()

			// Check result
			if got != tt.wantEnabled {
				t.Errorf("utils.IsColor() = %v, want %v", got, tt.wantEnabled)
			}
		})
	}
}

func TestColorBuiltins(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		isBuiltin bool
	}{
		{
			name:      "cd command",
			command:   "cd",
			isBuiltin: true,
		},
		{
			name:      "pwd command",
			command:   "pwd",
			isBuiltin: true,
		},
		{
			name:      "non-builtin command",
			command:   "ls",
			isBuiltin: false,
		},
		{
			name:      "empty command",
			command:   "",
			isBuiltin: false,
		},
		{
			name:      "export command",
			command:   "export",
			isBuiltin: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isBuiltin := utils.LinuxBuiltins[tt.command]
			if isBuiltin != tt.isBuiltin {
				t.Errorf("utils.LinuxBuiltins[%s] = %v, want %v",
					tt.command, isBuiltin, tt.isBuiltin)
			}
		})
	}
}

// TestColorFormatting tests the actual visual formatting of text
func TestColorFormatting(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		formats  []string
		expected string
	}{
		{
			name:     "bold red text",
			text:     "error",
			formats:  []string{utils.Bold, utils.TextRed},
			expected: utils.Bold + utils.TextRed + "error" + utils.Reset,
		},
		{
			name:     "underlined blue text with white background",
			text:     "info",
			formats:  []string{utils.Underline, utils.TextBlue, utils.BgWhite},
			expected: utils.Underline + utils.TextBlue + utils.BgWhite + "info" + utils.Reset,
		},
		{
			name:     "blinking yellow text",
			text:     "warning",
			formats:  []string{utils.Blink, utils.TextYellow},
			expected: utils.Blink + utils.TextYellow + "warning" + utils.Reset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ColorText(tt.text, tt.formats...)
			if result != tt.expected {
				t.Errorf("ColorText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestColorutils.Reset ensures that utils.Reset properly clears all formatting
func TestColorReset(t *testing.T) {
	formatted := utils.ColorText("test", utils.Bold+utils.TextRed+utils.BgWhite)
	if formatted[len(formatted)-len(utils.Reset):] != utils.Reset {
		t.Error("utils.Reset sequence not properly applied at the end of formatted text")
	}
}

// Mock shell state for testing
type MockShell struct {
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	env    map[string]string
}

func newMockShell() *MockShell {
	return &MockShell{
		stdout: new(bytes.Buffer),
		stderr: new(bytes.Buffer),
		env:    make(map[string]string),
	}
}

func (m *MockShell) setEnv(key, value string) {
	m.env[key] = value
}

func (m *MockShell) unsetEnv(key string) {
	delete(m.env, key)
}

// HandleColorCommand handles the color command and its arguments
func HandleColorCommand(args []string, shell *MockShell) error {
	if len(args) != 1 {
		shell.stderr.WriteString("Usage: color <on|off>\n")
		return ErrInvalidArgument
	}

	switch args[0] {
	case "on":
		shell.setEnv("SHELLCOLOR", "1")
		shell.stdout.WriteString("Color output enabled\n")
	case "off":
		shell.unsetEnv("SHELLCOLOR")
		shell.stdout.WriteString("Color output disabled\n")
	default:
		shell.stderr.WriteString("Invalid argument. Use 'on' or 'off'\n")
		return ErrInvalidArgument
	}

	return nil
}

// Test cases
func TestHandleColorCommand(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedOut  string
		expectedErr  string
		expectError  bool
		checkEnvVar  bool
		expectEnvVar bool
	}{
		{
			name:         "color on",
			args:         []string{"on"},
			expectedOut:  "Color output enabled\n",
			expectedErr:  "",
			expectError:  false,
			checkEnvVar:  true,
			expectEnvVar: true,
		},
		{
			name:         "color off",
			args:         []string{"off"},
			expectedOut:  "Color output disabled\n",
			expectedErr:  "",
			expectError:  false,
			checkEnvVar:  true,
			expectEnvVar: false,
		},
		{
			name:         "invalid argument",
			args:         []string{"invalid"},
			expectedOut:  "",
			expectedErr:  "Invalid argument. Use 'on' or 'off'\n",
			expectError:  true,
			checkEnvVar:  false,
			expectEnvVar: false,
		},
		{
			name:         "no arguments",
			args:         []string{},
			expectedOut:  "",
			expectedErr:  "Usage: color <on|off>\n",
			expectError:  true,
			checkEnvVar:  false,
			expectEnvVar: false,
		},
		{
			name:         "too many arguments",
			args:         []string{"on", "extra"},
			expectedOut:  "",
			expectedErr:  "Usage: color <on|off>\n",
			expectError:  true,
			checkEnvVar:  false,
			expectEnvVar: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shell := newMockShell()

			err := HandleColorCommand(tt.args, shell)

			// Check error
			if (err != nil) != tt.expectError {
				t.Errorf("HandleColorCommand() error = %v, expectError %v",
					err, tt.expectError)
			}

			// Check stdout
			if got := shell.stdout.String(); got != tt.expectedOut {
				t.Errorf("stdout = %q, want %q", got, tt.expectedOut)
			}

			// Check stderr
			if got := shell.stderr.String(); got != tt.expectedErr {
				t.Errorf("stderr = %q, want %q", got, tt.expectedErr)
			}

			// Check environment variable
			if tt.checkEnvVar {
				_, exists := shell.env["SHELLCOLOR"]
				if exists != tt.expectEnvVar {
					t.Errorf("SHELLCOLOR env var exists = %v, want %v",
						exists, tt.expectEnvVar)
				}
			}
		})
	}
}

// TestColorCommandSequence tests a sequence of color commands
func TestColorCommandSequence(t *testing.T) {
	shell := newMockShell()

	// Test sequence: on -> on -> off -> off
	sequences := []struct {
		args         []string
		expectEnvVar bool
	}{
		{[]string{"on"}, true},
		{[]string{"on"}, true}, // Should still be on
		{[]string{"off"}, false},
		{[]string{"off"}, false}, // Should still be off
	}

	for i, seq := range sequences {
		shell.stdout.Reset()
		shell.stderr.Reset()

		err := HandleColorCommand(seq.args, shell)
		if err != nil {
			t.Errorf("Step %d: unexpected error: %v", i, err)
		}

		_, exists := shell.env["SHELLCOLOR"]
		if exists != seq.expectEnvVar {
			t.Errorf("Step %d: SHELLCOLOR exists = %v, want %v",
				i, exists, seq.expectEnvVar)
		}
	}
}

// TestColorCommandWithEmptyEnvironment tests color command behavior with empty environment
func TestColorCommandWithEmptyEnvironment(t *testing.T) {
	shell := newMockShell()

	// First check initial state
	if utils.IsColor() {
		t.Error("Color should be disabled in initial state")
	}

	// Enable color
	err := HandleColorCommand([]string{"on"}, shell)
	if err != nil {
		t.Errorf("Unexpected error enabling color: %v", err)
	}

	// Check if color is enabled
	_, exists := shell.env["SHELLCOLOR"]
	if !exists {
		t.Error("SHELLCOLOR should be set after enabling")
	}

	// Disable color
	err = HandleColorCommand([]string{"off"}, shell)
	if err != nil {
		t.Errorf("Unexpected error disabling color: %v", err)
	}

	// Check if color is disabled
	_, exists = shell.env["SHELLCOLOR"]
	if exists {
		t.Error("SHELLCOLOR should not be set after disabling")
	}
}

// TestColorCommandCaseSensitivity tests case sensitivity of arguments
func TestColorCommandCaseSensitivity(t *testing.T) {
	tests := []struct {
		arg           string
		shouldSucceed bool
	}{
		{"on", true},
		{"ON", false},
		{"On", false},
		{"off", true},
		{"OFF", false},
		{"Off", false},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			shell := newMockShell()
			err := HandleColorCommand([]string{tt.arg}, shell)
			if (err == nil) != tt.shouldSucceed {
				t.Errorf("HandleColorCommand(%q) success = %v, want %v",
					tt.arg, err == nil, tt.shouldSucceed)
			}
		})
	}
}
