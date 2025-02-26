package exit

import (
	user "asa/shell/internal/service"
	"asa/shell/utils"
	"fmt"
	"io"
	"os"
	"strconv"

	"gorm.io/gorm"
)

type ExitCommand struct {
	user *user.User
	db   *gorm.DB
}

func NewExitCommand(db *gorm.DB, user *user.User) *ExitCommand {
	return &ExitCommand{
		user: user,
		db:   db,
	}
}

func (c *ExitCommand) Name() string {
	return "exit"
}

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
			return utils.ErrInvalidArgs
		}
		fmt.Fprintln(stdout, "exit status ", status)
		os.Exit(0)
	default:
		return utils.ErrTooManyArgs
	}
	return nil
}
