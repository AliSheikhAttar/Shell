package shell

import (
	"asa/shell/internal/command"
	"asa/shell/utils"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var (
	ErrCommandNotSupported = errors.New("command not found")
)

type Shell struct {
	reader   *bufio.Reader
	commands map[string]command.Command
	rootDir  string
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
}

// New creates a new shell instance
func New() *Shell {
	sh := &Shell{
		reader:   bufio.NewReader(os.Stdin),
		commands: make(map[string]command.Command),
		stdin:    os.Stdin,  // Default to standard input
		stdout:   os.Stdout, // Default to standard output
		stderr:   os.Stderr, // Default to standard error
	}
	// Define built-in commands
	builtins := []string{"exit", "echo", "cat", "type", "cd"} // Add all built-in commands here

	// Register the exit command
	exitCmd := command.NewExitCommand()
	sh.registerCommand(exitCmd)

	// Register the echo command
	echoCmd := command.NewEchoCommand()
	sh.registerCommand(echoCmd)

	// Register the cat command
	catCmd := command.NewCatCommand()
	sh.registerCommand(catCmd)

	// Register the type command
	typeCmd := command.NewTypeCommand(builtins)
	sh.registerCommand(typeCmd)

	// Register the pwd command
	pwdCmd := command.NewPwdCommand()
	sh.registerCommand(pwdCmd)

	// Register the cd command
	cdCmd := command.NewCDCommand(sh.rootDir)
	sh.registerCommand(cdCmd)

	// Register the ls command
	lsCmd := command.NewLSCommand()
	sh.commands[lsCmd.Name()] = lsCmd	

	stdout := &bytes.Buffer{}
	sh.commands["pwd"].Execute([]string{}, stdout)
	sh.rootDir = stdout.String()

	return sh
}

func (s *Shell) registerCommand(cmd command.Command) {
	s.commands[cmd.Name()] = cmd
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
			fmt.Fprintf(s.stderr, "%s: %v\n", input, err)
			if err == ErrCommandNotSupported {
				fmt.Println()
				fmt.Fprintln(s.stdout, "List of supported builtin commands are as followings: ")
				for key := range s.commands {
					fmt.Fprintln(s.stdout, key)
				}
			}
		}
	}
}

func (s *Shell) executeSystemCommand(name string, args []string) error {
	// Use type command to find the executable path
	execPath, err := utils.FindCommand(name)
	if err != nil {
		return err
	}
	if utils.HasPrefix(execPath, "$builtin") {
		execPath = strings.Split(execPath, ":")[1] // seperate builtin command
	}

	// Create and execute the system command with the full path
	cmd := exec.Command(execPath, args...)

	// Use Shell's IO streams instead of os package defaults
	cmd.Stdin = s.stdin
	cmd.Stdout = s.stdout
	cmd.Stderr = s.stderr

	// Execute the command
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute %s: %v", name, err)
	}

	return nil
}

// printPrompt displays the shell prompt
func (s *Shell) printPrompt() error {
	currentDir, err := utils.CurrentPwd()
	if err != nil {
		return err
	}
	addr := utils.HandleAdress(s.rootDir, currentDir)

	_, err = fmt.Fprintf(s.stdout, "%s$ ", addr)
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
		err := command.Execute(args, s.stdout)
		if err != nil {
			return err
		}
		return nil
	}

	// if err := s.executeSystemCommand(cmd, args); err != nil {
	// 	return ErrCommandNotSupported
	// }

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
