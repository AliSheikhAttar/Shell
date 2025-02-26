package shell

import (
	"asa/shell/internal/command"
	"asa/shell/internal/command/adduser"
	"asa/shell/internal/command/cat"
	"asa/shell/internal/command/cd"
	"asa/shell/internal/command/color"
	"asa/shell/internal/command/echo"
	"asa/shell/internal/command/exit"
	"asa/shell/internal/command/help"
	"asa/shell/internal/command/history"
	"asa/shell/internal/command/login"
	"asa/shell/internal/command/logout"
	"asa/shell/internal/command/ls"
	"asa/shell/internal/command/pwd"
	typecmd "asa/shell/internal/command/type"
	db "asa/shell/internal/database"
	"asa/shell/internal/redirection"
	user "asa/shell/internal/service"
	"asa/shell/utils"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrCommandNotSupported = errors.New("command not found")
	ErrNotValidDirectory   = errors.New("current directory is not valid")
)

type Shell struct {
	reader   *bufio.Reader
	user     user.User
	database *gorm.DB
	commands map[string]command.Command
	history  map[string]int
	rootDir  string
}

type std struct {
	std          *os.File
	isRedirected bool
}
type redirect struct {
	stdout    *std
	stderr    *std
	redirType redirection.RedirectionType
}

func New() *Shell {
	rootDir, err := utils.CurrentPwd()
	if err != nil {
		return &Shell{}
	}

	err = db.GetDB().AutoMigrate(&user.User{})
	if err != nil {
		fmt.Println("Error migrating database:", err)
	}


	sh := &Shell{
		user:     user.User{Username: ""},
		database: db.GetDB(),
		reader:   bufio.NewReader(os.Stdin),
		commands: make(map[string]command.Command),
		history:  make(map[string]int),
		rootDir:  rootDir,
	}
	exitCmd := exit.NewExitCommand(sh.database, &sh.user)
	sh.registerCommand(exitCmd)

	echoCmd := echo.NewEchoCommand()
	sh.registerCommand(echoCmd)

	catCmd := cat.NewCatCommand()
	sh.registerCommand(catCmd)

	pwdCmd := pwd.NewPwdCommand()
	sh.registerCommand(pwdCmd)

	cdCmd := cd.NewCDCommand(sh.rootDir)
	sh.registerCommand(cdCmd)

	lsCmd := ls.NewLSCommand()
	sh.commands[lsCmd.Name()] = lsCmd

	colorCmd := color.NewColorCommand()
	sh.commands[colorCmd.Name()] = colorCmd

	loginCmd := login.NewLoginCommand(sh.database, &sh.user)
	sh.commands[loginCmd.Name()] = loginCmd

	adduserCmd := adduser.NewAddUserCommand(sh.database, &sh.user)
	sh.commands[adduserCmd.Name()] = adduserCmd

	logoutCmd := logout.NewLogoutCommand(sh.database, &sh.user)
	sh.commands[logoutCmd.Name()] = logoutCmd

	historyCmd := history.NewHistoryCommand(&sh.history, &sh.user, sh.database)
	sh.commands[historyCmd.Name()] = historyCmd

	helpCmd := help.NewHelpCommand()
	sh.commands[helpCmd.Name()] = helpCmd

	shellBuiltins := []string{}
	for cmd := range sh.commands {
		shellBuiltins = append(shellBuiltins, cmd)
	}
	typeCmd := typecmd.NewTypeCommand(shellBuiltins)
	sh.registerCommand(typeCmd)

	stdout := &bytes.Buffer{}
	sh.commands["pwd"].Execute([]string{}, stdout)
	sh.rootDir = stdout.String()

	// if err := utils.ClearAndFillHistoryWithMockData(sh.database); err != nil {
	// 	log.Fatalf("Error clearing and filling history: %v", err)
	// }
	return sh
}

func (s *Shell) registerCommand(cmd command.Command) {
	s.commands[cmd.Name()] = cmd
}

func (s *Shell) Start() error {
	for {
		if err := s.printPrompt(); err != nil {
			return err
		}
		input, err := s.readInput()
		if err != nil {
			return err
		}
		if input == "" {
			continue
		}
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
	user := s.user.Username
	if utils.IsColor() {
		addr = utils.ColorText(addr, utils.TextBlue)
		user = utils.ColorText(s.user.Username, utils.TextGreen)
	}
	if s.user.Username != "" {
		_, err = fmt.Fprintf(os.Stdout, "%s:%s$ ", user, addr)
	} else {
		_, err = fmt.Fprintf(os.Stdout, "%s$ ", addr)
	}
	return err
}

func (s *Shell) readInput() (string, error) {
	input, err := s.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func (s *Shell) executeCommand(input string) (*std, error) {
	cmd, args, redirects, err := s.parseCommand(input)
	if cmd != "history" {
		if s.user.Username != "" {
			s.user.HistoryMap[input]++
			// err := user.Update(s.database, &s.user) // too insufficient but most reliable
		} else {
			s.history[input]++
		}
	}

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

	if err := s.executeSystemCommand(cmd, args, redirects.stdout.std, redirects.stderr.std); err != nil {
		return redirects.stderr, ErrCommandNotSupported
	}

	return redirects.stderr, nil
}

func (s *Shell) executeSystemCommand(name string, args []string, stdout io.Writer, stderr io.Writer) error {
	execPath, err := utils.FindCommand(name)
	if err != nil {
		return err
	}
	if utils.HasPrefix(execPath, "$builtin") {
		execPath = strings.Split(execPath, ":")[1]
	}

	cmd := exec.Command(execPath, args...)

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute %s: %v", name, err)
	}

	return nil
}

func (s *Shell) parseCommand(input string) (string, []string, *redirect, error) {
	redirects := &redirect{stdout: &std{os.Stdout, false}, stderr: &std{os.Stderr, false}}
	parsedArg, err1 := utils.ParseArgs(input)
	if err1 != nil {
		return "", nil, redirects, nil
	}
	if len(parsedArg) == 0 {
		return "", nil, redirects, nil
	}
	// Parse redirection
	args, redir, err := redirection.ParseRedirection(parsedArg)
	if err != nil {
		return "", nil, redirects, err
	}
	// Setup redirection if needed
	if redir != nil {
		file, err := redirection.SetupRedirection(redir)
		if err != nil {
			return "", nil, redirects, err
		}
		// Set appropriate output & error
		switch redir.Type {
		case redirection.OutputRedirect, redirection.OutputAppend:
			redirects.stdout.std = file
			redirects.stdout.isRedirected = true
		case redirection.ErrorRedirect, redirection.ErrorAppend:
			redirects.stderr.std = file
			redirects.stderr.isRedirected = true
		}
		redirects.redirType = redir.Type
	}
	// for case : > file3 cat file2
	if parsedArg[0][0] != '>' {
		return parsedArg[0], args[1:], redirects, err1
	} else {
		return parsedArg[2], args, redirects, err1
	}
}
