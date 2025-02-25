package history

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
		return utils.ErrInvalidArgs
	}
	if len(args) == 1 {
		if args[0] != "clean" {
			return utils.ErrInvalidArgs
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
		if len(h.user.HistoryMap) == 0 {
			return utils.ErrEmptyHistory
		}
		utils.PrintSortedMap(h.user.HistoryMap, stdout)
		return nil
	}
	if len(*h.builtinHistory) == 0 {
		return utils.ErrEmptyHistory
	}
	utils.PrintSortedMap(*h.builtinHistory, stdout)
	return nil
}
