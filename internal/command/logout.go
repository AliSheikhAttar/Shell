package command

import (
	user "asa/shell/internal/service"
	"io"

	"gorm.io/gorm"
)

type LogoutCommand struct {
	db   *gorm.DB
	user *user.User
}

func NewLogoutCommand(db *gorm.DB, user *user.User) *LogoutCommand {
	return &LogoutCommand{
		db:   db,
		user: user,
	}
}

func (c *LogoutCommand) Name() string {
	return "logout"
}

func (c *LogoutCommand) Execute(args []string, stdout io.Writer) error {
	if len(args) > 1 {
		return ErrInvalidArgs
	}

	err := user.Update(c.db, c.user)
	if err != nil {
		return err
	}
	(*c.user).ID = 0
	(*c.user).Username = ""
	(*c.user).Password = ""

	return nil
}
