package ls

import (
	"bytes"

	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLSCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedOutput string
		wantErr     bool
		setup       func(tempDir string) error
	}{
		{
			name:        "no arguments - list current directory",
			args:        []string{},
			expectedOutput: "dir1\nfile1.txt\nfile2.txt",
			wantErr:     false,
			setup: func(tempDir string) error {
				err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content1"), 0644)
				if err != nil {
					return err
				}
				err = os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("content2"), 0644)
				if err != nil {
					return err
				}
				err = os.Mkdir(filepath.Join(tempDir, "dir1"), 0755)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name:        "list specific directory",
			args:        []string{"subdir"},
			expectedOutput: "subfile1.txt\n",
			wantErr:     false,
			setup: func(tempDir string) error {
				subdirPath := filepath.Join(tempDir, "subdir")
				err := os.Mkdir(subdirPath, 0755)
				if err != nil {
					return err
				}
				err = os.WriteFile(filepath.Join(subdirPath, "subfile1.txt"), []byte("subcontent1"), 0644)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name:        "list with -a flag - show all files",
			args:        []string{"-a"},
			expectedOutput: func() string {
				return ".hidden_file\ndir1\nfile1.txt\n"
			}(),
			wantErr:     false,
			setup: func(tempDir string) error {
				err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content1"), 0644)
				if err != nil {
					return err
				}
				err = os.Mkdir(filepath.Join(tempDir, "dir1"), 0755)
				if err != nil {
					return err
				}
				err = os.WriteFile(filepath.Join(tempDir, ".hidden_file"), []byte("hidden content"), 0644)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name:        "list with -l flag - long format",
			args:        []string{"-l"},
			expectedOutput: func() string {
				// The content of long format is dynamic (time, size), so we check for presence of parts
				return `drwxr-xr-x     4096 * dir1
-rw-r--r--        8 * file1.txt
-rw-r--r--        8 * file2.txt
						`
			}(),
			wantErr:     false,
			setup: func(tempDir string) error {
				err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content1"), 0644)
				if err != nil {
					return err
				}
				err = os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("content2"), 0644)
				if err != nil {
					return err
				}
				err = os.Mkdir(filepath.Join(tempDir, "dir1"), 0755)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name:        "list with -al flags - all and long format",
			args:        []string{"-al"},
			expectedOutput: func() string {
				return `-rw-r--r--       14 * .hidden_file
drwxr-xr-x     4096 * dir1
-rw-r--r--        8 * file1.txt

`
			}(),
			wantErr:     false,
			setup: func(tempDir string) error {
				err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content1"), 0644)
				if err != nil {
					return err
				}
				err = os.Mkdir(filepath.Join(tempDir, "dir1"), 0755)
				if err != nil {
					return err
				}
				err = os.WriteFile(filepath.Join(tempDir, ".hidden_file"), []byte("hidden content"), 0644)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			name:    "invalid option",
			args:    []string{"-x"},
			wantErr: true,
		},
		{
			name:    "too many arguments",
			args:    []string{"dir1", "dir2"},
			wantErr: true,
		},
		{
			name:    "directory not found",
			args:    []string{"non_existent_dir"},
			wantErr: true,
		},
		{
			name:        "empty directory",
			args:        []string{"empty_dir"},
			expectedOutput: "",
			wantErr:     false,
			setup: func(tempDir string) error {
				emptyDirPath := filepath.Join(tempDir, "empty_dir")
				return os.Mkdir(emptyDirPath, 0755)
			},
		},
		{
			name:        "list hidden directory with -a",
			args:        []string{"-a", ".hidden_dir"},
			expectedOutput: "sub_hidden_file\n",
			wantErr:     false,
			setup: func(tempDir string) error {
				hiddenDirPath := filepath.Join(tempDir, ".hidden_dir")
				err := os.Mkdir(hiddenDirPath, 0755)
				if err != nil {
					return err
				}
				err = os.WriteFile(filepath.Join(hiddenDirPath, "sub_hidden_file"), []byte("hidden content"), 0644)
				if err != nil {
					return err
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewLSCommand()
			tempDir, err := os.MkdirTemp("", "ls-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}
			if err := os.Chdir(tempDir); err != nil {
				t.Fatalf("Failed to chdir to temp dir: %v", err)
			}
			defer os.Chdir(originalDir) 

			currentDir, _ := os.Getwd() 
			t.Logf("Test '%s' - Current working directory: %s", tt.name, currentDir)
			t.Logf("Test '%s' - Temp directory: %s", tt.name, tempDir)


			if tt.setup != nil {
				if err := tt.setup(tempDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			for i, arg := range tt.args {
				if arg == "subdir" {
					tt.args[i] = filepath.Join(tempDir, "subdir")
				}
				if arg == "empty_dir" {
					tt.args[i] = filepath.Join(tempDir, "empty_dir")
				}
				if arg == ".hidden_dir" {
					tt.args[i] = filepath.Join(tempDir, ".hidden_dir")
				}
			}


			stdout := &bytes.Buffer{}
			err = cmd.Execute(tt.args, stdout)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			gotOutput := strings.TrimSpace(stdout.String())
			expectedOutput := strings.TrimSpace(tt.expectedOutput)

			if tt.name == "list with -l flag - long format" || tt.name == "list with -al flags - all and long format" || tt.name == "list with -a flag - show all files"{ // Apply pattern matching for -a and -l tests
				expectedLines := strings.Split(strings.TrimSpace(tt.expectedOutput), "\n")
				gotLines := strings.Split(gotOutput, "\n")

				if len(gotLines) != len(expectedLines) {
					t.Errorf("LSCommand.Execute() output line count = %d, want %d", len(gotLines), len(expectedLines))
					t.Logf("Got output:\n%s\nExpected output:\n%s", gotOutput, tt.expectedOutput) 
					return
				}
				for i := range expectedLines {
					matched, _ := filepath.Match(expectedLines[i], gotLines[i]) 
					if !matched {
						t.Errorf("LSCommand.Execute() output line %d = \n%v\n, want to match pattern \n%v", i, gotLines[i], expectedLines[i])
						t.Logf("Got output:\n%s\nExpected output:\n%s", gotOutput, tt.expectedOutput)
						return
					}
				}


			} else {
				if gotOutput != expectedOutput {
					t.Errorf("LSCommand.Execute() output = \n%v\n, want \n%v", gotOutput, expectedOutput)
				}
			}
		})
	}
}