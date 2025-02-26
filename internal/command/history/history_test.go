package history

import (
	db "asa/shell/internal/database"
	userSvc "asa/shell/internal/service"
	"asa/shell/utils"
	"bytes"
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestHistoryCommand_Execute(t *testing.T) {
	tests := []struct {
		name                  string
		args                  []string
		builtinHistory        *map[string]int
		user                  *userSvc.User
		db                    *gorm.DB
		wantErr               error
		builtinHistoryCleaned bool
	}{
		{
			name:    "Too many arguments",
			args:    []string{"arg1", "arg2"},
			wantErr: utils.ErrInvalidArgs,
		},
		{
			name:    "Invalid argument",
			args:    []string{"invalid_arg"},
			wantErr: utils.ErrInvalidArgs,
		},
		{
			name:           "Clean builtin history",
			args:           []string{"clean"},
			builtinHistory: &map[string]int{"cmd1": 1, "cmd2": 2},
			user: &userSvc.User{
				Username:   "",
				HistoryMap: map[string]int{"cmd1": 1, "cmd2": 2},
			},
			wantErr:               nil,
			builtinHistoryCleaned: true,
		},
		{
			name: "Clean user history",
			args: []string{"clean"},
			user: &userSvc.User{
				Username:   "testuser",
				HistoryMap: map[string]int{"cmd1": 1, "cmd2": 2},
			},
			db:      db.GetDB(),
			wantErr: nil,
		},
		{
			name: "Clean user history - Update User Error",
			args: []string{"clean"},
			user: &userSvc.User{
				Username:   "testuser",
				HistoryMap: map[string]int{"cmd1": 1, "cmd2": 2},
			},
			db:      db.GetDB(),
			wantErr: nil,
		},
		{
			name: "Empty user history",
			args: []string{},
			user: &userSvc.User{
				Username:   "testuser",
				HistoryMap: map[string]int{},
			},
			wantErr: utils.ErrEmptyHistory,
		},
		{
			name: "Non-empty user history",
			args: []string{},
			user: &userSvc.User{
				Username:   "testuser",
				HistoryMap: map[string]int{"cmd1": 1, "cmd2": 2, "cmd3": 3},
			},
			wantErr: nil,
		},
		{
			name:           "Empty builtin history",
			args:           []string{},
			builtinHistory: &map[string]int{},
			wantErr:        utils.ErrEmptyHistory,
		},
		{
			name:           "Non-empty builtin history",
			args:           []string{},
			builtinHistory: &map[string]int{"cmd1": 1, "cmd2": 2, "cmd3": 3},
			wantErr:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builtinHistory := tt.builtinHistory
			if builtinHistory == nil {
				builtinHistory = &map[string]int{}
			}
			builtinHistoryCopy := make(map[string]int)
			for k, v := range *builtinHistory {
				builtinHistoryCopy[k] = v
			}
			h := &HistoryCommand{
				builtinHistory: &builtinHistoryCopy,
				user:           tt.user,
				db:             tt.db,
			}
			var stdout bytes.Buffer
			builtinHistoryInitial := make(map[string]int)
			if tt.builtinHistory != nil {
				for k, v := range *tt.builtinHistory {
					builtinHistoryInitial[k] = v
				}
			}

			hDb := db.GetDB()
			h.db = hDb
			var testUser userSvc.User
			testUser, err := userSvc.GetUser(hDb, "testuser", "")
			if err == nil {
				testUser.HistoryMap = make(map[string]int)
			} else {
				testUser = userSvc.User{Username: "testuser"}
				userSvc.RegisterUser(hDb, &testUser)
				testUser, _ = userSvc.GetUser(hDb, "testuser", "")
			}
			if tt.user == nil || tt.user.Username == "" {
				testUser.Username = ""
			}
			h.user = &testUser
			if tt.user != nil && len(tt.user.HistoryMap) != 0 {
				h.user.HistoryMap = tt.user.HistoryMap
			}

			err = h.Execute(tt.args, &stdout)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.builtinHistoryCleaned {
				if len(*h.builtinHistory) != 0 {
					t.Errorf("Execute() builtinHistory should be cleaned but is not, got %v, initial was %v", *h.builtinHistory, builtinHistoryInitial)
				}
			}
		})
	}
}

func TestHistoryCommand_Name(t *testing.T) {
	h := NewHistoryCommand(nil, nil, nil)
	if h.Name() != "history" {
		t.Errorf("Name() should return 'history', got %v", h.Name())
	}
}
