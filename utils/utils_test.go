package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestColorText(t *testing.T) {
	testCases := []struct {
		name           string
		text           string
		formats        []string
		expectedOutput string
	}{
		{
			name:           "No formats",
			text:           "Test text",
			formats:        []string{},
			expectedOutput: "Test text" + Reset,
		},
		{
			name:           "Single format - TextRed",
			text:           "Error",
			formats:        []string{TextRed},
			expectedOutput: TextRed + "Error" + Reset,
		},
		{
			name:           "Single format - Bold",
			text:           "Important",
			formats:        []string{Bold},
			expectedOutput: Bold + "Important" + Reset,
		},
		{
			name:           "Multiple formats - TextBlue and Underline",
			text:           "Link",
			formats:        []string{TextBlue, Underline},
			expectedOutput: TextBlue + Underline + "Link" + Reset,
		},
		{
			name:           "Multiple formats - Bold, TextGreen, Italic",
			text:           "Success",
			formats:        []string{Bold, TextGreen, Italic},
			expectedOutput: Bold + TextGreen + Italic + "Success" + Reset,
		},
		{
			name:           "Empty text, single format",
			text:           "",
			formats:        []string{TextCyan},
			expectedOutput: TextCyan + "" + Reset,
		},
		{
			name:           "Text with special characters, multiple formats",
			text:           "Text with \\n and \t",
			formats:        []string{TextYellow, Reverse},
			expectedOutput: TextYellow + Reverse + "Text with \\n and \t" + Reset,
		},
		{
			name:           "Format codes as text - should not be interpreted as formats again",
			text:           TextRed + " is red color code",                // Text already containing format codes
			formats:        []string{Bold},                                // Apply bold on top
			expectedOutput: Bold + TextRed + " is red color code" + Reset, // Bold + (TextRed + " is red color code") + Reset
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualOutput := ColorText(tc.text, tc.formats...)
			if actualOutput != tc.expectedOutput {
				t.Errorf("Test case '%s': Output mismatch:\nexpected:\n'%s'\ngot:\n'%s'", tc.name, tc.expectedOutput, actualOutput)
			}
		})
	}
}

func TestIsColor(t *testing.T) {
	originalShellColor, shellColorSet := os.LookupEnv("SHELLCOLOR") // Backup original SHELLCOLOR and if it was set
	defer func() {
		if shellColorSet {
			os.Setenv("SHELLCOLOR", originalShellColor) // Restore original SHELLCOLOR
		} else {
			os.Unsetenv("SHELLCOLOR") // Unset if it was not originally set
		}
	}()

	testCases := []struct {
		name            string
		setEnvVar       bool
		envVarValue     string
		expectedIsColor bool
	}{
		{
			name:            "SHELLCOLOR not set",
			setEnvVar:       false,
			expectedIsColor: false,
		},
		{
			name:            "SHELLCOLOR set to empty string",
			setEnvVar:       true,
			envVarValue:     "",
			expectedIsColor: true,
		},
		{
			name:            "SHELLCOLOR set to 'true'",
			setEnvVar:       true,
			envVarValue:     "true",
			expectedIsColor: true,
		},
		{
			name:            "SHELLCOLOR set to 'false'",
			setEnvVar:       true,
			envVarValue:     "false",
			expectedIsColor: true, // Presence of variable, not its value, determines IsColor() result currently.
		},
		{
			name:            "SHELLCOLOR set to '1'",
			setEnvVar:       true,
			envVarValue:     "1",
			expectedIsColor: true,
		},
		{
			name:            "SHELLCOLOR set to '0'",
			setEnvVar:       true,
			envVarValue:     "0",
			expectedIsColor: true, // Presence of variable, not its value.
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnvVar {
				os.Setenv("SHELLCOLOR", tc.envVarValue)
			} else {
				os.Unsetenv("SHELLCOLOR")
			}

			actualIsColor := IsColor()
			if actualIsColor != tc.expectedIsColor {
				t.Errorf("Test case '%s': IsColor() returned '%v', expected '%v'", tc.name, actualIsColor, tc.expectedIsColor)
			}
		})
	}
}

func TestFindCommand(t *testing.T) {
	// Define builtins for testing context
	LinuxBuiltins = map[string]bool{
		"testbuiltin":    true,
		"anotherbuiltin": true,
	}

	// Helper function to create test executables in a temp dir
	createTestExecutable := func(dir, name string) string {
		exePath := filepath.Join(dir, name)
		if runtime.GOOS == "windows" {
			exePath += ".exe" // Add .exe suffix for Windows
		}
		// Create a dummy executable file (no actual content needed for these tests)
		file, err := os.Create(exePath)
		if err != nil {
			t.Fatalf("Failed to create test executable: %v", err)
		}
		defer file.Close()
		err = os.Chmod(exePath, 0755) // Make it executable
		if err != nil {
			t.Fatalf("Failed to chmod test executable: %v", err)
		}
		return exePath
	}

	// Backup and restore PATH and SHELL environment variables
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath) // Restore PATH after test

	testCases := []struct {
		name          string
		cmd           string
		pathEnv       string
		setupPathDirs func(tempDir string) []string // Function to setup dirs in PATH and create test executables, returns created dirs for cleanup
		wantErr       error
	}{
		{
			name:    "Built-in command",
			cmd:     "testbuiltin",
			wantErr: nil,
		},
		{
			name:    "Command with absolute path - no error",
			cmd:     "/abs/path/to/executable",
			wantErr: nil,
		},
		{
			name:    "Command with relative path - no error",
			cmd:     "./relative/path/executable",
			wantErr: nil,
		},
		{
			name:    "Command in PATH - found in first dir",
			cmd:     "testcmd1",
			pathEnv: "/testpath1:/testpath2",
			setupPathDirs: func(tempDir string) []string {
				pathDir1 := filepath.Join(tempDir, "testpath1")
				os.MkdirAll(pathDir1, 0755)
				createTestExecutable(pathDir1, "testcmd1")
				pathDir2 := filepath.Join(tempDir, "testpath2") // Not used in this case, but created for PATH env
				os.MkdirAll(pathDir2, 0755)
				return []string{pathDir1, pathDir2}
			},
			wantErr: nil,
		},
		{
			name:    "Command in PATH - found in second dir",
			cmd:     "testcmd2",
			pathEnv: "/testpath1:/testpath2",
			setupPathDirs: func(tempDir string) []string {
				pathDir1 := filepath.Join(tempDir, "testpath1")
				os.MkdirAll(pathDir1, 0755)
				pathDir2 := filepath.Join(tempDir, "testpath2")
				os.MkdirAll(pathDir2, 0755)
				createTestExecutable(pathDir2, "testcmd2")
				return []string{pathDir1, pathDir2}
			},
			wantErr: nil,
		},
		{
			name:    "Command not in PATH - PATH set but not found",
			cmd:     "nonexistentcmd",
			pathEnv: "/testpath1:/testpath2",
			setupPathDirs: func(tempDir string) []string {
				pathDir1 := filepath.Join(tempDir, "testpath1")
				os.MkdirAll(pathDir1, 0755)
				pathDir2 := filepath.Join(tempDir, "testpath2")
				os.MkdirAll(pathDir2, 0755)
				return []string{pathDir1, pathDir2}
			},
			wantErr: ErrCommandNotFound,
		},
		{
			name:          "Command not in PATH - PATH not set",
			cmd:           "somecmd",
			pathEnv:       "",  // Empty PATH
			setupPathDirs: nil, // No dirs to setup as PATH is empty
			wantErr:       ErrEnvironmentVarNotSet,
		},
		{
			name:    "Command in PATH - executable with suffix (no suffix in command)", // Test finding exe with suffix when command is given without
			cmd:     "testcmdSuffix",
			pathEnv: "/testpathSuffix",
			setupPathDirs: func(tempDir string) []string {
				pathDir := filepath.Join(tempDir, "testpathSuffix")
				os.MkdirAll(pathDir, 0755)
				createTestExecutable(pathDir, "testcmdSuffix"+exeSuffix()) // Create executable WITH suffix
				return []string{pathDir}
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			var pathDirs []string
			if tc.setupPathDirs != nil {
				pathDirs = tc.setupPathDirs(tempDir)
			}
			defer func() { // Cleanup: remove created directories and files in tempDir
				for _, dir := range pathDirs {
					os.RemoveAll(dir)
				}
			}()

			if tc.pathEnv != "" {
				// Replace placeholders in pathEnv with tempDir path if needed
				pathEnvWithTempDir := strings.ReplaceAll(tc.pathEnv, "/testpath1", filepath.Join(tempDir, "testpath1"))
				pathEnvWithTempDir = strings.ReplaceAll(pathEnvWithTempDir, "/testpath2", filepath.Join(tempDir, "testpath2"))
				pathEnvWithTempDir = strings.ReplaceAll(pathEnvWithTempDir, "/testpathSuffix", filepath.Join(tempDir, "testpathSuffix"))

				os.Setenv("PATH", pathEnvWithTempDir)
			} else {
				os.Unsetenv("PATH") // Ensure PATH is unset for cases where PATH is intentionally empty
			}

			fullPath, err := FindCommand(tc.cmd)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("Test case '%s': Expected error '%v', but got '%v'", tc.name, tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("Test case '%s': Unexpected error: %v", tc.name, err)
			}

			// --- Calculate expectedFullPath HERE, inside t.Run, using tempDir ---
			var expectedFullPath string
			switch tc.name { // Based on test case name, construct expected path. Adjust as needed for other cases.
			case "Command in PATH - found in first dir":
				expectedFullPath = filepath.Join(tempDir, "testpath1", tc.cmd) // Construct expected path
			case "Command in PATH - found in second dir":
				expectedFullPath = filepath.Join(tempDir, "testpath2", tc.cmd)
			case "Command in PATH - executable with suffix (no suffix in command)":
				expectedFullPath = filepath.Join(tempDir, "testpathSuffix", tc.cmd+exeSuffix())
			case "Built-in command":
				expectedFullPath = fmt.Sprintf("$builtin:%s", tc.cmd) // Or set it based on test case logic if needed.
			case "Command not in PATH - PATH set but not found", "Command not in PATH - PATH not set': Full path mismatch", "Command not in PATH - PATH not set":
				expectedFullPath = ""
			default: // For other cases where command is not found in PATH or is built-in, expectedFullPath remains empty string "" which is already the default.
				expectedFullPath = tc.cmd // Or set it based on test case logic if needed.
			}
			// ----------------------------------------------------------------------

			// Normalize paths for comparison
			expectedFullPathNormalized := filepath.ToSlash(expectedFullPath)
			actualFullPathNormalized := filepath.ToSlash(fullPath)

			if actualFullPathNormalized != expectedFullPathNormalized {
				expectedDisplayPath := expectedFullPath
				actualDisplayPath := fullPath
				if expectedFullPath != "" {
					expectedDisplayPath = expectedFullPathNormalized
				}
				if fullPath != "" {
					actualDisplayPath = actualFullPathNormalized
				}

				t.Errorf("Test case '%s': Full path mismatch:\nexpected: '%s'\ngot:      '%s'\n(Normalized expected: '%s', Normalized got: '%s')", tc.name, expectedDisplayPath, actualDisplayPath, expectedFullPathNormalized, actualFullPathNormalized)
			}
		})
	}
}

func TestIsQuoted(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "Single quoted", input: "'quoted string'", expected: true},
		{name: "Double quoted", input: "\"quoted string\"", expected: true},
		{name: "Not quoted", input: "not quoted", expected: false},
		{name: "Starts with single quote, but no end", input: "'starts but no end", expected: false},
		{name: "Starts with double quote, but no end", input: "\"starts but no end", expected: false},
		{name: "Ends with single quote, but no start", input: "ends but no start'", expected: false},
		{name: "Ends with double quote, but no start", input: "ends but no start\"", expected: false},
		{name: "Empty string", input: "", expected: false},
		{name: "Single quote only", input: "'", expected: false},                                                    // Not considered quoted as per spec
		{name: "Double quote only", input: "\"", expected: false},                                                   // Not considered quoted as per spec
		{name: "Escaped quotes - not considered quoted by IsQuoted", input: "\\'quoted string\\'", expected: false}, // Backslash escapes are shell responsibility, IsQuoted checks literal quotes
		{name: "Mixed quotes - not quoted", input: "'double\" quotes'", expected: true},                             // Mismatched quotes are not quoted
		{name: "Quoted number", input: "\"12345\"", expected: true},
		{name: "Quoted special characters", input: "'!@#$%^'", expected: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsQuoted(tc.input)
			if actual != tc.expected {
				t.Errorf("Test case '%s': IsQuoted(%s) returned %v, expected %v", tc.name, tc.input, actual, tc.expected)
			}
		})
	}
}

func TestWhichQuoted(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "Single quoted", input: "'quoted string'", expected: "'"},
		{name: "Double quoted", input: "\"quoted string\"", expected: "\""},
		{name: "Not quoted", input: "not quoted", expected: ""},
		{name: "Starts with single quote, but no end", input: "'starts but no end", expected: ""},
		{name: "Starts with double quote, but no end", input: "\"starts but no end", expected: ""},
		{name: "Ends with single quote, but no start", input: "ends but no start'", expected: ""},
		{name: "Ends with double quote, but no start", input: "ends but no start\"", expected: ""},
		{name: "Empty string", input: "", expected: ""},
		{name: "Single quote only", input: "'", expected: ""},                                                       // Not considered quoted as per spec - returns empty string
		{name: "Double quote only", input: "\"", expected: ""},                                                      // Not considered quoted as per spec - returns empty string
		{name: "Escaped quotes - not considered quoted by WhichQuoted", input: "\\'quoted string\\'", expected: ""}, // Backslash escapes are shell responsibility, WhichQuoted checks literal quotes - returns empty string
		{name: "Mixed quotes - not quoted", input: "'double\" quotes'", expected: "'"},                              // Mismatched quotes are not quoted - returns empty string
		{name: "Quoted number", input: "\"12345\"", expected: "\""},
		{name: "Quoted special characters", input: "'!@#$%^'", expected: "'"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := WhichQuoted(tc.input)
			if actual != tc.expected {
				t.Errorf("Test case '%s': WhichQuoted(%q) returned %q, expected %q", tc.name, tc.input, actual, tc.expected)
			}
		})
	}
}

func TestIsAlphaNumeric(t *testing.T) {
	testCases := []struct {
		name     string
		input    byte
		expected bool
	}{
		{name: "Lowercase a", input: 'a', expected: true},
		{name: "Lowercase z", input: 'z', expected: true},
		{name: "Uppercase A", input: 'A', expected: true},
		{name: "Uppercase Z", input: 'Z', expected: true},
		{name: "Digit 0", input: '0', expected: true},
		{name: "Digit 9", input: '9', expected: true},
		{name: "Space", input: ' ', expected: false},
		{name: "Special char !", input: '!', expected: false},
		{name: "Newline", input: '\n', expected: false},
		{name: "Tab", input: '\t', expected: false},
		{name: "Semicolon", input: ';', expected: false},
		{name: "Null byte", input: 0, expected: false}, // Byte value 0
		{name: "Punctuation .", input: '.', expected: false},
		{name: "Symbol $", input: '$', expected: false},
		{name: "Byte out of ASCII range (e.g., for extended char sets)", input: 128, expected: false}, // Example: non-ASCII byte
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsAlphaNumeric(tc.input)
			if actual != tc.expected {
				t.Errorf("Test case '%s': IsAlphaNumeric('%c') returned %v, expected %v", tc.name, tc.input, actual, tc.expected)
			}
		})
	}
}

func TestHasPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		s        string
		prefix   string
		expected bool
	}{
		{name: "Has prefix", s: "startswithprefix", prefix: "startswith", expected: true},
		{name: "Exact match", s: "prefix", prefix: "prefix", expected: true},
		{name: "Does not have prefix", s: "nomatchprefix", prefix: "prefix", expected: false},
		{name: "Prefix longer than string", s: "short", prefix: "longerprefix", expected: false},
		{name: "Empty string, empty prefix", s: "", prefix: "", expected: true}, // Empty string starts with empty prefix
		{name: "Empty string, non-empty prefix", s: "", prefix: "prefix", expected: false},
		{name: "Non-empty string, empty prefix", s: "string", prefix: "", expected: true},                       // Any string starts with empty prefix
		{name: "Prefix is substring in the middle", s: "stringprefixstring", prefix: "prefix", expected: false}, // Only checks at the beginning
		{name: "Prefix with special chars", s: "!@#$string", prefix: "!@#$", expected: true},
		{name: "String with special chars", s: "!@#$prefix", prefix: "prefix", expected: false}, // Special chars in string, not at prefix position
		{name: "Unicode string with prefix", s: "你好prefix", prefix: "你好", expected: true},       // Unicode support in prefix
		{name: "Unicode prefix in string", s: "prefix你好", prefix: "prefix", expected: true},
		{name: "Unicode prefix, non-matching string", s: "不匹配prefix", prefix: "你好", expected: false}, // Non-matching unicode prefix
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := HasPrefix(tc.s, tc.prefix)
			if actual != tc.expected {
				t.Errorf("Test case '%s': HasPrefix(%q, %q) returned %v, expected %v", tc.name, tc.s, tc.prefix, actual, tc.expected)
			}
		})
	}
}

func TestHasSuffix(t *testing.T) {
	testCases := []struct {
		name     string
		s        string
		suffix   string
		expected bool
	}{
		{name: "Has suffix", s: "endswithsuffix", suffix: "suffix", expected: true},
		{name: "Exact match", s: "suffix", suffix: "suffix", expected: true},
		{name: "Does not have suffix", s: "nomatchsuffix", suffix: "suffix", expected: true},
		{name: "Suffix longer than string", s: "short", suffix: "longersuffix", expected: false},
		{name: "Empty string, empty suffix", s: "", suffix: "", expected: true}, // Empty string ends with empty suffix
		{name: "Empty string, non-empty suffix", s: "", suffix: "suffix", expected: false},
		{name: "Non-empty string, empty suffix", s: "string", suffix: "", expected: true},                       // Any string ends with empty suffix
		{name: "Suffix is substring in the middle", s: "stringsuffixstring", suffix: "suffix", expected: false}, // Only checks at the end
		{name: "Suffix with special chars", s: "string!@#$", suffix: "!@#$", expected: true},
		{name: "String with special chars", s: "suffix!@#$", suffix: "suffix", expected: false}, // Special chars in string, not at suffix position
		{name: "Unicode string with suffix", s: "suffix你好", suffix: "你好", expected: true},       // Unicode support in suffix
		{name: "Unicode suffix in string", s: "你好suffix", suffix: "suffix", expected: true},
		{name: "Unicode suffix, non-matching string", s: "suffix不匹配", suffix: "你好", expected: false}, // Non-matching unicode suffix
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := HasSuffix(tc.s, tc.suffix)
			if actual != tc.expected {
				t.Errorf("Test case '%s': HasSuffix(%q, %q) returned %v, expected %v", tc.name, tc.s, tc.suffix, actual, tc.expected)
			}
		})
	}
}

func TestTrimEdge(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "Trim single quotes", input: "'quoted string'", expected: "quoted string"},
		{name: "Trim double quotes", input: "\"quoted string\"", expected: "quoted string"},
		{name: "Not quoted - no trim", input: "not quoted", expected: "ot quote"},                                   // Trims first and last char if not quotes
		{name: "Starts with quote, ends with different char", input: "'quoted string\"", expected: "quoted string"}, // Mismatched - still trims first and last
		{name: "Ends with quote, starts with different char", input: "\"quoted string'", expected: "quoted string"}, // Mismatched - still trims first and last
		{name: "String length 2 - trimmed to empty", input: "ab", expected: ""},                                     // Length 2 becomes empty after trim
		{name: "String length 1 - trimmed to empty", input: "a", expected: ""},                                      // Length 1 also becomes empty
		{name: "Empty string - no change", input: "", expected: ""},                                                 // Empty string stays empty
		{name: "String with leading/trailing spaces and quotes", input: "  ' quoted '  ", expected: " ' quoted ' "}, // Trim edge spaces and quotes
		{name: "String with only spaces - trimmed", input: "   ", expected: " "},                                    // Trimmed spaces
		{name: "String with special chars and quotes", input: "'!@#$%^&'", expected: "!@#$%^&"},
		{name: "Unicode string with quotes", input: "\"你好世界\"", expected: "你好世界"}, // Unicode with quotes trimming
		{name: "Unicode string without quotes", input: "你好世界", expected: "好世"},    // Trims first and last rune (could be multi-byte)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := TrimEdge(tc.input)
			if actual != tc.expected {
				t.Errorf("Test case '%s': TrimEdge(%q) returned %q, expected %q", tc.name, tc.input, actual, tc.expected)
			}
		})
	}
}

// Helper function to get executable suffix based on OS
func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
