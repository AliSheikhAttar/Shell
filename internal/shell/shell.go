package shell

import (
	"asa/shell/internal/command"
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Shell struct {
	reader   *bufio.Reader
	commands map[string]command.Command // Changed from exitCmd to commands map
}

// New creates a new shell instance
func New() *Shell {
	sh := &Shell{
		reader:   bufio.NewReader(os.Stdin),
		commands: make(map[string]command.Command),
	}

	// Register the exit command
	exitCmd := command.NewExitCommand()
	sh.commands[exitCmd.Name()] = exitCmd

	return sh
}

// Start begins the shell's read-eval-print loop
func (s *Shell) Start() error {
	for {
		if err := s.printPrompt(); err != nil {
			return err
		}

		// read inpug line
		input, err := s.readInput()
		if err != nil {
			return err
		}

		// Handle empty input
		if input == "" {
			continue
		}

		// Process the command (for now, just echo it back)
		if err := s.executeCommand(input); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}
}

// printPrompt displays the shell prompt
func (s *Shell) printPrompt() error {
	_, err := fmt.Print("$ ")
	return err
}

// readInput reads a line of input from the user
func (s *Shell) readInput() (string, error) {
	input, err := s.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Trim whitespace and newline
	return strings.TrimSpace(input), nil
}

// executeCommand processes the input command (currently just echoes it)
func (s *Shell) executeCommand(input string) error {
	cmd, args := s.parseCommand(input)

	if command, exists := s.commands[cmd]; exists {
		err := command.Execute(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return nil
		}
		return nil
	}

	// For now, echo other commands
	fmt.Println("You entered:", input)
	return nil
}

// parseCommand splits the input into command and arguments
func (s *Shell) parseCommand(input string) (string, []string) {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return "", nil
	}
	return fields[0], fields[1:]
}

// ... rest of the shell.go implementation remains the same ...
