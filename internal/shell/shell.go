package shell

import (
	"asa/shell/internal/command"
	"asa/shell/internal/redirection"
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

var linuxBuiltins = map[string]bool{
	"cd":      true,
	"pwd":     true,
	"exit":    true,
	"echo":    true,
	"export":  true,
	"source":  true,
	"alias":   true,
	"unalias": true,
	"set":     true,
	"unset":   true,
	"exec":    true,
	"command": true,
	".":       true,
}

type Shell struct {
	reader   *bufio.Reader
	commands map[string]command.Command
	rootDir  string
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
}

type std struct {
	std          *os.File
	isRedirected bool
}
type redirect struct {
	stdout *std
	stderr *std
}

// New creates a new shell instance
func New() *Shell {
	sh := &Shell{
		reader:   bufio.NewReader(os.Stdin),
		commands: make(map[string]command.Command),
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
		if stderr, err := s.executeCommand(input); err != nil {
			if stderr.isRedirected {
				defer stderr.std.Close()
			}

			fmt.Fprintf(stderr.std, "%s: %v\n", input, err)
			if err == ErrCommandNotSupported {
				fmt.Println()
				fmt.Fprintln(stderr.std, "List of supported builtin commands are as followings: ")
				for key := range s.commands {
					fmt.Fprintln(stderr.std, key)
				}
			}
		}
	}
}

// printPrompt displays the shell prompt
func (s *Shell) printPrompt() error {
	currentDir, err := utils.CurrentPwd()
	if err != nil {
		return err
	}
	addr := utils.HandleAdress(s.rootDir, currentDir)

	_, err = fmt.Fprintf(os.Stdout, "%s$ ", addr)
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
func (s *Shell) executeCommand(input string) (*std, error) {
	cmd, args, redirects, err := s.parseCommand(input)
	if err != nil {
		return redirects.stderr, err
	}

	if redirects.stdout.isRedirected {

		defer redirects.stdout.std.Close()
	}

	if command, exists := s.commands[cmd]; exists {
		err := command.Execute(args, redirects.stdout.std)
		if err != nil {
			return redirects.stderr, err
		}
		return redirects.stderr, nil
	}

	// linux builtin command not implemented
	if _, isExist := linuxBuiltins[cmd]; isExist {
		return redirects.stderr, ErrCommandNotSupported
	}

	// system commands
	if err := s.executeSystemCommand(cmd, args); err != nil {
		return redirects.stderr, ErrCommandNotSupported
	}

	return redirects.stderr, nil
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

	// Execute the command
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute %s: %v", name, err)
	}

	return nil
}

// parseCommand splits the input into command and arguments
func (s *Shell) parseCommand(input string) (string, []string, *redirect, error) {
	fields := strings.Fields(input)
	redirects := &redirect{stdout: &std{os.Stdout, false}, stderr: &std{os.Stderr, false}}

	if len(fields) == 0 {
		return "", nil, redirects, nil
	}

	// Parse redirection
	args, redir, err := redirection.ParseRedirection(fields[1:])
	if err != nil {
		return "", nil, redirects, err
	}
	// Setup redirection if needed
	if redir != nil {
		file, err := redirection.SetupRedirection(redir)
		if err != nil {
			return "", nil, redirects, err
		}
		// Set appropriate output
		switch redir.Type {
		case redirection.OutputRedirect, redirection.OutputAppend:
			redirects.stdout.std = file
			redirects.stdout.isRedirected = true
		case redirection.ErrorRedirect, redirection.ErrorAppend:
			redirects.stderr.std = file
			redirects.stderr.isRedirected = true
		}
	}
	return fields[0], args, redirects, nil
}
