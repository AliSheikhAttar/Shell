package command

import (
	user "asa/shell/internal/service"
	"io"

	"gorm.io/gorm"
)

type AddUserCommand struct {
	db   *gorm.DB
	user *user.User
}


func NewAddUserCommand(db *gorm.DB, user *user.User) *AddUserCommand {
	return &AddUserCommand{
		db:   db,
		user: user,
	}
}

func (c *AddUserCommand) Name() string {
	return "adduser"
}

func (c *AddUserCommand) Execute(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return ErrUsernameRequired
	}
	if len(args) > 2 {
		return ErrInvalidArgs
	}
	var pass string
	if len(args) == 2 {
		pass = args[1]
	}

	newUser := &user.User{Username: args[0], Password: pass}
	err := user.RegisterUser(c.db, newUser)
	if err != nil {
		return err
	}
	return nil
}
