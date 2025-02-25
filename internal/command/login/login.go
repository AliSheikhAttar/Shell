package login

import (
	userService "asa/shell/internal/service"
	"asa/shell/utils"
	"io"

	"gorm.io/gorm"
)

type LoginCommand struct {
	db   *gorm.DB
	user *userService.User
}

func NewLoginCommand(db *gorm.DB, user *userService.User) *LoginCommand {
	return &LoginCommand{
		db:   db,
		user: user,
	}
}

func (c *LoginCommand) Name() string {
	return "login"
}

func (c *LoginCommand) Execute(args []string, stdout io.Writer) error {
	// if (*c.user).Username != "" {
	// 	return ErrLoggedin
	// }
	if len(args) == 0 {
		return utils.ErrUsernameRequired
	}
	if len(args) > 2 {
		return utils.ErrInvalidArgs
	}
	var pass string
	if len(args) == 2 {
		pass = args[1]
	}
	user, err := userService.GetUser(c.db, args[0], pass)
	if err != nil {
		return err
	}
	if c.user.Username != "" {
		err := userService.Update(c.db, c.user)
		if err != nil {
			return err
		}
	}
	(*c.user).ID = user.ID
	(*c.user).Username = user.Username
	(*c.user).Password = user.Password
	(*c.user).History = user.History
	(*c.user).HistoryMap = user.HistoryMap

	return nil
}
