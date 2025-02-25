package cat

import (
    "os"
    "path/filepath"
    "testing"
)

func TestCatCommand(t *testing.T) {
    // Create temporary test files
    tmpDir := t.TempDir()
    
    // Test file 1
    testFile1 := filepath.Join(tmpDir, "test1.txt")
    content1 := "Hello\nWorld\n"
    if err := os.WriteFile(testFile1, []byte(content1), 0644); err != nil {
        t.Fatal(err)
    }

    // Test file 2
    testFile2 := filepath.Join(tmpDir, "test2.txt")
    content2 := "Testing\nCat\nCommand\n"
    if err := os.WriteFile(testFile2, []byte(content2), 0644); err != nil {
        t.Fatal(err)
    }

    tests := []struct {
        name    string
        args    []string
        wantErr bool
    }{
        {
            name:    "no arguments",
            args:    []string{},
            wantErr: true,
        },
        {
            name:    "single file",
            args:    []string{testFile1},
            wantErr: false,
        },
		{
            name:    "file in quote",
            args:    []string{testFile1},
            wantErr: false,
        },
        {
            name:    "multiple files",
            args:    []string{testFile1, testFile2},
            wantErr: false,
        },
        {
            name:    "non-existent file",
            args:    []string{"nonexistent.txt"},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := NewCatCommand()
            
            // Verify command name
            if got := cmd.Name(); got != "cat" {
                t.Errorf("CatCommand.Name() = %v, want %v", got, "cat")
            }

            // Execute command
            err := cmd.Execute(tt.args, os.Stdout)
            if (err != nil) != tt.wantErr {
                t.Errorf("CatCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}