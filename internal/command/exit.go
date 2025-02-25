package command

import (
	user "asa/shell/internal/service"
	"fmt"
	"io"
	"os"
	"strconv"

	"gorm.io/gorm"
)

// ExitCommand implements the 'exit' built-in command
type ExitCommand struct {
	user *user.User
	db   *gorm.DB
}

// NewExitCommand creates a new exit command
func NewExitCommand(db *gorm.DB, user *user.User) *ExitCommand {
	return &ExitCommand{
		user: user,
		db:   db,
	}
}

// Name returns the name of the command
func (c *ExitCommand) Name() string {
	return "exit"
}

// Execute handles the exit command execution
func (c *ExitCommand) Execute(args []string, stdout io.Writer) error {

	switch len(args) {
	case 0:
		if c.user.Username != "" {
			user.Update(c.db, c.user)
		}
		fmt.Fprintln(stdout, "exit status 0")
		os.Exit(0)

	case 1:
		if c.user.Username != "" {
			user.Update(c.db, c.user)
		}
		status, err := strconv.Atoi(args[0])
		if err != nil {
			return ErrInvalidArgs
		}
		fmt.Fprintln(stdout, "exit status ", status)
		os.Exit(0)
	default:
		return ErrTooManyArgs
	}
	return nil
}
