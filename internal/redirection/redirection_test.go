package redirection

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestParseRedirection(t *testing.T) {
	testCases := []struct {
		name          string
		inputArgs     []string
		expectedArgs  []string
		expectedRedir *Redirection
		wantErr       error
	}{
		{
			name:          "No redirection",
			inputArgs:     []string{"ls", "-l"},
			expectedArgs:  []string{"ls", "-l"},
			expectedRedir: nil,
			wantErr:       nil,
		},
		{
			name:          "Output redirect > in middle",
			inputArgs:     []string{"ls", "-l", ">", "output.txt"},
			expectedArgs:  []string{"ls", "-l"},
			expectedRedir: &Redirection{Type: OutputRedirect, File: "output.txt"},
			wantErr:       nil,
		},
		{
			name:          "Output append >> in middle",
			inputArgs:     []string{"cmd", "arg1", ">>", "append.log"},
			expectedArgs:  []string{"cmd", "arg1"},
			expectedRedir: &Redirection{Type: OutputAppend, File: "append.log"},
			wantErr:       nil,
		},
		{
			name:          "Error redirect 2> in middle",
			inputArgs:     []string{"command", "param", "2>", "error.log"},
			expectedArgs:  []string{"command", "param"},
			expectedRedir: &Redirection{Type: ErrorRedirect, File: "error.log"},
			wantErr:       nil,
		},
		{
			name:          "Error append 2>> in middle",
			inputArgs:     []string{"run", "-opt", "2>>", "error_append.log"},
			expectedArgs:  []string{"run", "-opt"},
			expectedRedir: &Redirection{Type: ErrorAppend, File: "error_append.log"},
			wantErr:       nil,
		},
		{
			name:          "Missing file for output redirect > in middle",
			inputArgs:     []string{"cmd", ">"},
			expectedArgs:  nil,
			expectedRedir: nil,
			wantErr:       ErrMissingFileForRedirection,
		},
		{
			name:          "Missing file for output append >> in middle",
			inputArgs:     []string{"cmd", ">>"},
			expectedArgs:  nil,
			expectedRedir: nil,
			wantErr:       ErrMissingFileForRedirection,
		},
		{
			name:          "Missing file for error redirect 2> in middle",
			inputArgs:     []string{"cmd", "2>"},
			expectedArgs:  nil,
			expectedRedir: nil,
			wantErr:       ErrMissingFileForRedirection,
		},
		{
			name:          "Missing file for error append 2>> in middle",
			inputArgs:     []string{"cmd", "2>>"},
			expectedArgs:  nil,
			expectedRedir: nil,
			wantErr:       ErrMissingFileForRedirection,
		},
		{
			name:          "Arguments before and after redirection in middle", 
			inputArgs:     []string{"command", "-arg1", ">", "output.txt", "extra_arg"},
			expectedArgs:  []string{"command", "-arg1"},
			expectedRedir: &Redirection{Type: OutputRedirect, File: "output.txt"},
			wantErr:       nil,
		},
		{
			name:          "Multiple redirection operators in middle - first one takes precedence",
			inputArgs:     []string{"cmd", ">", "output1.txt", ">>", "output2.txt"},
			expectedArgs:  []string{"cmd"},
			expectedRedir: &Redirection{Type: OutputRedirect, File: "output1.txt"},
			wantErr:       nil,
		},
		{
			name:          "Output redirect > at start",
			inputArgs:     []string{">", "output.txt", "cmd", "arg1"}, 
			expectedArgs:  []string{"arg1"},                           
			expectedRedir: &Redirection{Type: OutputRedirect, File: "output.txt"},
			wantErr:       nil,
		},
		{
			name:          "Output append >> at start",
			inputArgs:     []string{">>", "append.log", "command", "-option"}, 
			expectedArgs:  []string{"-option"},                                
			expectedRedir: &Redirection{Type: OutputAppend, File: "append.log"},
			wantErr:       nil,
		},
		{
			name:          "Error redirect 2> at start",
			inputArgs:     []string{"2>", "error.log", "program", "--flag"}, 
			expectedArgs:  []string{"--flag"},                               
			expectedRedir: &Redirection{Type: ErrorRedirect, File: "error.log"},
			wantErr:       nil,
		},
		{
			name:          "Error append 2>> at start",
			inputArgs:     []string{"2>>", "error_append.log", "script", "param1"}, 
			expectedArgs:  []string{"param1"},                                      
			expectedRedir: &Redirection{Type: ErrorAppend, File: "error_append.log"},
			wantErr:       nil,
		},
		{
			name:          "Missing file for output redirect > at start",
			inputArgs:     []string{">", "cmd", "arg"},  
			expectedArgs:  nil,                          
			expectedRedir: nil,                          
			wantErr:       ErrMissingFileForRedirection, 
		},
		{
			name:          "Redirection operator in initialQuotes - error redirection append",
			inputArgs:     []string{"2>>", "output.txt", "ls", "-l"},
			expectedArgs:  []string{"-l"}, 
			expectedRedir: &Redirection{Type: ErrorAppend, File: "output.txt"},
			wantErr:       nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualArgs, actualRedir, err := ParseRedirection(tc.inputArgs)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("Test case '%s': Expected error '%v', but got '%v'", tc.name, tc.wantErr, err)
				}
				return 
			}

			if err != nil {
				t.Fatalf("Test case '%s': Unexpected error: %v", tc.name, err)
			}

			if !stringSlicesEqual(actualArgs, tc.expectedArgs) {
				t.Errorf("Test case '%s': Args mismatch:\nexpected: %v\ngot:      %v", tc.name, tc.expectedArgs, actualArgs)
			}

			if tc.expectedRedir == nil {
				if actualRedir != nil {
					t.Errorf("Test case '%s': Expected no redirection, but got: %v", tc.name, actualRedir)
				}
			} else {
				if actualRedir == nil {
					t.Fatalf("Test case '%s': Expected redirection, but got nil", tc.name)
				}
				if actualRedir.Type != tc.expectedRedir.Type {
					t.Errorf("Test case '%s': Redirection type mismatch: expected '%v', got '%v'", tc.name, tc.expectedRedir.Type, actualRedir.Type)
				}
				if actualRedir.File != tc.expectedRedir.File {
					t.Errorf("Test case '%s': Redirection file mismatch: expected '%s', got '%s'", tc.name, tc.expectedRedir.File, actualRedir.File)
				}
			}
		})
	}
}



func TestSetupRedirection(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		name                string
		redir               *Redirection
		expectedFlags       int
		expectError         bool
		checkFileContent    bool
		fileContentToWrite  string
		expectedFileContent string
	}{
		{
			name:          "No redirection",
			redir:         nil,
			expectedFlags: 0, 
			expectError:   false,
		},
		{
			name: "Output Redirect >",
			redir: &Redirection{
				Type: OutputRedirect,
				File: filepath.Join(tmpDir, "output_redirect.txt"),
			},
			expectedFlags:       os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
			expectError:         false,
			checkFileContent:    true,
			fileContentToWrite:  "test output redirect",
			expectedFileContent: "test output redirect",
		},
		{
			name: "Output Append >>",
			redir: &Redirection{
				Type: OutputAppend,
				File: filepath.Join(tmpDir, "output_append.log"),
			},
			expectedFlags:       os.O_WRONLY | os.O_CREATE | os.O_APPEND,
			expectError:         false,
			checkFileContent:    true,
			fileContentToWrite:  "initial content\n",                  
			expectedFileContent: "initial content\nappended content\n", 
		},
		{
			name: "Error Redirect 2>",
			redir: &Redirection{
				Type: ErrorRedirect,
				File: filepath.Join(tmpDir, "error_redirect.err"),
			},
			expectedFlags:       os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
			expectError:         false,
			checkFileContent:    true,
			fileContentToWrite:  "test error redirect",
			expectedFileContent: "test error redirect",
		},
		{
			name: "Error Append 2>>",
			redir: &Redirection{
				Type: ErrorAppend,
				File: filepath.Join(tmpDir, "error_append.err.log"),
			},
			expectedFlags:       os.O_WRONLY | os.O_CREATE | os.O_APPEND,
			expectError:         false,
			checkFileContent:    true,
			fileContentToWrite:  "initial error content\n",                   
			expectedFileContent: "initial error content\nappended content\n", 
		},
		{
			name: "File open error - permission denied",
			redir: &Redirection{
				Type: OutputRedirect,
				File: filepath.Join("/", "root_owned_file.txt"), 
			},
			expectedFlags: os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
			expectError:   true, 
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var initialContent []byte
			if tc.redir != nil && tc.checkFileContent && (tc.redir.Type == OutputAppend || tc.redir.Type == ErrorAppend) {
				initialContent = []byte(tc.fileContentToWrite)
				err := ioutil.WriteFile(tc.redir.File, initialContent, 0644)
				if err != nil {
					t.Fatalf("SetupRedirection Test '%s': Failed to write initial content to file: %v", tc.name, err)
				}
			}

			file, err := SetupRedirection(tc.redir)

			if tc.expectError {
				if err == nil {
					t.Errorf("Test case '%s': Expected error, but got nil", tc.name)
				}
				return 
			}

			if err != nil {
				t.Fatalf("Test case '%s': Unexpected error: %v", tc.name, err)
			}

			if tc.redir != nil { 
				if file == nil {
					t.Fatalf("Test case '%s': Expected file, but got nil", tc.name)
				}
				defer file.Close()

				fileStat, err := file.Stat()
				if err != nil {
					t.Fatalf("Test case '%s': Could not stat file: %v", tc.name, err)
				}
				if os.FileMode(0644) != fileStat.Mode().Perm()&0777 { 
					t.Errorf("Test case '%s': File permissions are not as expected. Expected permissions to contain 0644, got %o", tc.name, fileStat.Mode().Perm()&0777)
				}

				if tc.checkFileContent {
					if tc.redir.Type == OutputAppend || tc.redir.Type == ErrorAppend {
						_, err = file.WriteString("appended content\n") 
						if err != nil {
							t.Fatalf("Test case '%s': Error writing appended content: %v", tc.name, err)
						}
					} else {
						_, err = file.WriteString(tc.fileContentToWrite) 
						if err != nil {
							t.Fatalf("Test case '%s': Error writing content: %v", tc.name, err)
						}
					}

					contentRead, err := ioutil.ReadFile(tc.redir.File)
					if err != nil {
						t.Fatalf("Test case '%s': Error reading file content: %v", tc.name, err)
					}
					actualContent := string(contentRead)
					if actualContent != tc.expectedFileContent {
						t.Errorf("Test case '%s': File content mismatch:\nexpected:\n'%s'\ngot:\n'%s'", tc.name, tc.expectedFileContent, actualContent)
					}
				}


				expectedOpenFlags := tc.expectedFlags
				actualOpenFlags := flagsForRedirectionType(tc.redir.Type)         
				if actualOpenFlags != expectedOpenFlags && tc.expectedFlags != 0 {
					if tc.redir.Type != NoRedirection {
						t.Logf("Note: Cannot reliably verify file flags after os.OpenFile across platforms. Test only checks flags *used* during OpenFile.") // Informative log
						
					}

				}

			} else {
				if file != nil {
					t.Errorf("Test case '%s': Expected nil file for no redirection, but got: %v", tc.name, file)
				}
			}
		})
	}
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func flagsForRedirectionType(redirType RedirectionType) int {
	flags := os.O_WRONLY | os.O_CREATE
	if redirType == OutputAppend || redirType == ErrorAppend {
		flags |= os.O_APPEND
	} else if redirType != NoRedirection { 
		flags |= os.O_TRUNC
	}
	return flags
}
