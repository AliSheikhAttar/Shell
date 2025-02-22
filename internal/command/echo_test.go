package command

import (
    "os"
    "testing"
)

func TestEchoCommand(t *testing.T) {
    // Set up test environment variable
    os.Setenv("PATH1", "/usr/local/bin:/usr/bin:/bin")
    os.Setenv("HOME1", "/home/user")

    tests := []struct {
        name     string
        args     []string
        expected string
    }{
        {
            name:     "simple text",
            args:     []string{"hello"},
            expected: "hello\n",
        },
        {
            name:     "multiple words",
            args:     []string{"hello", "world"},
            expected: "hello world\n",
        },
        {
            name:     "with environment variable",
            args:     []string{"Path1:", "$PATH"},
            expected: "Path1: /usr/local/bin:/usr/bin:/bin\n",
        },
        {
            name:     "quoted text",
            args:     []string{"'hello $PATH'"},
            expected: "hello $PATH\n",
        },
        {
            name:     "mixed content",
            args:     []string{"Home1", "is", "$HOME"},
            expected: "Home1 is /home/user\n",
        },
        {
            name:     "environment variable in word",
            args:     []string{"no$PATH1"},
            expected: "no/usr/local/bin:/usr/bin:/bin\n",
        },
		{
            name:     "quoted environment variable",
            args:     []string{"$'PATH'"},
            expected: "PATHn",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := NewEchoCommand(os.Stdout)
            
            // Verify command name
            if got := cmd.Name(); got != "echo" {
                t.Errorf("EchoCommand.Name() = %v, want %v", got, "echo")
            }

            // Capture stdout
            // Note: In a real implementation, you might want to use a more sophisticated
            // way to capture stdout, but this is simplified for the example
            err := cmd.Execute(tt.args)
            if err != nil {
                t.Errorf("EchoCommand.Execute() error = %v", err)
            }
        })
    }
}