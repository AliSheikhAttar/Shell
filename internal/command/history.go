package command

import (
	user "asa/shell/internal/service"
	"asa/shell/utils"
	"io"

	"gorm.io/gorm"
)

type HistoryCommand struct {
	builtinHistory *map[string]int
	user           *user.User
	db             *gorm.DB
}

func NewHistoryCommand(builtinHistory *map[string]int, user *user.User, db *gorm.DB) *HistoryCommand {
	return &HistoryCommand{
		builtinHistory: builtinHistory,
		user:           user,
		db:             db,
	}
}

func (h *HistoryCommand) Name() string {
	return "history"
}

func (h *HistoryCommand) Execute(args []string, stdout io.Writer) error {
	if len(args) > 1 {
		return ErrInvalidArgs
	}
	if len(args) == 1 {
		if args[0] != "clean" {
			return ErrInvalidArgs
		}
		if h.user.Username != "" {
			h.user.HistoryMap = map[string]int{}
			err := user.Update(h.db, h.user)
			return err
		}
		*h.builtinHistory = map[string]int{}
		return nil
	}
	if h.user.Username != "" {
		utils.PrintSortedMap(h.user.HistoryMap, stdout)
		return nil
	}

	utils.PrintSortedMap(*h.builtinHistory, stdout)
	return nil
}
