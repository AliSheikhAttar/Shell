package typecmd

import (
    "os"
    "path/filepath"
    "testing"
)

func TestTypeCommand(t *testing.T) {
    // Create temporary test directory and add it to PATH
    tmpDir := t.TempDir()
    originalPath := os.Getenv("PATH")
    os.Setenv("PATH", tmpDir+":"+originalPath)
    defer os.Setenv("PATH", originalPath)

    // Create a test executable in the temporary directory
    testExec := filepath.Join(tmpDir, "testcmd")
    if err := os.WriteFile(testExec, []byte("#!/bin/sh\necho test"), 0755); err != nil {
        t.Fatal(err)
    }

    // Define built-in commands
    builtins := []string{"cd", "echo", "exit", "type", "cat"}

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
            name:    "built-in command",
            args:    []string{"cd"},
            wantErr: false,
        },
        {
            name:    "executable command",
            args:    []string{"testcmd"},
            wantErr: false,
        },
        {
            name:    "non-existent command",
            args:    []string{"nonexistentcmd"},
            wantErr: true,
        },
        {
            name:    "multiple commands",
            args:    []string{"cd", "testcmd"},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := NewTypeCommand(builtins)
            
            // Verify command name
            if got := cmd.Name(); got != "type" {
                t.Errorf("TypeCommand.Name() = %v, want %v", got, "type")
            }

            // Execute command
            err := cmd.Execute(tt.args, os.Stdout)
            if (err != nil) != tt.wantErr {
                t.Errorf("TypeCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

