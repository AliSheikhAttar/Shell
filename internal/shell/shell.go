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
	// Register the exit command
	exitCmd := command.NewExitCommand()
	sh.registerCommand(exitCmd)

	// Register the echo command
	echoCmd := command.NewEchoCommand()
	sh.registerCommand(echoCmd)

	// Register the cat command
	catCmd := command.NewCatCommand()
	sh.registerCommand(catCmd)

	// Register the pwd command
	pwdCmd := command.NewPwdCommand()
	sh.registerCommand(pwdCmd)

	// Register the cd command
	cdCmd := command.NewCDCommand(sh.rootDir)
	sh.registerCommand(cdCmd)

	// Register the ls command
	lsCmd := command.NewLSCommand()
	sh.commands[lsCmd.Name()] = lsCmd

	// Register the color command
	colorCmd := command.NewColorCommand()
	sh.commands[colorCmd.Name()] = colorCmd
	// Register the type command

	shellBuiltins := []string{}
	for cmd := range sh.commands {
		shellBuiltins = append(shellBuiltins, cmd)
	}
	typeCmd := command.NewTypeCommand(shellBuiltins)
	sh.registerCommand(typeCmd)

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

		// Process the command and write errors
		if stderr, err := s.executeCommand(input); err != nil {
			if stderr.isRedirected {
				defer stderr.std.Close()
			}
			cmdError := fmt.Sprintf("%s: %v\n", input, err)
			if utils.IsColor() {
				cmdError = utils.ColorText(cmdError, utils.TextRed)
			}
			fmt.Fprintf(stderr.std, "%s", cmdError)
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

	currendDir := fmt.Sprintf("%s$ ", addr)
	if utils.IsColor() {
		currendDir = utils.ColorText(currendDir, utils.TextBlue)
	}
	_, err = fmt.Fprintf(os.Stdout, "%s", currendDir)
	return err
}

// readInput reads a line of input from the user
func (s *Shell) readInput() (string, error) {
	input, err := s.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	// input := "echo 'hello \"mom\"    world'  hello \"hi    'dad'   there    !\""
	// input := "echo 'home'"
	// input := "echo 'hello    \"world\" is nice ' sds     jlsd  \"      x is       here     \" ok"
	// input := "echo 'hello    \"world\" is nice ' sds     jlsd  \"      x is       here     \" ok\""
	// input := "echo '\"hello\"'"
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
	if _, isExist := utils.LinuxBuiltins[cmd]; isExist {
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
	redirects := &redirect{stdout: &std{os.Stdout, false}, stderr: &std{os.Stderr, false}}
	quotes, err1 := utils.ExtractQuotes(input)
	fields := utils.Seperate(input, quotes)

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
	return fields[0], args, redirects, err1
}
