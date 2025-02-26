package history

// import (
// 	"bytes"
// 	"testing"

// 	user "asa/shell/internal/service"
// 	"asa/shell/utils"

// 	"gorm.io/gorm"
// )

// func TestHistoryCommand_Execute(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		args           []string
// 		builtinHistory map[string]int
// 		setupUser      func(db *gorm.DB) *user.User
// 		expectedErr    error
// 		expectedOutput string
// 		assertHistory  func(t *testing.T, cmd *HistoryCommand, db *gorm.DB) // For clean command checks
// 	}{
// 		{
// 			name:           "Display builtin history - not empty",
// 			args:           []string{},
// 			builtinHistory: map[string]int{"ls": 2, "pwd": 1},
// 			setupUser: func(db *gorm.DB) *user.User {
// 				return &user.User{} // No logged-in user (Username is empty)
// 			},
// 			expectedErr:    nil,
// 			expectedOutput: "ls: 2\npwd: 1\n",
// 			assertHistory:  func(t *testing.T, cmd *HistoryCommand, db *gorm.DB) {},
// 		},
// 		{
// 			name:           "Display builtin history - empty",
// 			args:           []string{},
// 			builtinHistory: map[string]int{},
// 			setupUser: func(db *gorm.DB) *user.User {
// 				return &user.User{} // No logged-in user
// 			},
// 			expectedErr:    utils.ErrEmptyHistory,
// 			expectedOutput: "",
// 			assertHistory:  func(t *testing.T, cmd *HistoryCommand, db *gorm.DB) {},
// 		},
// 		{
// 			name: "Display user history - not empty",
// 			args: []string{},
// 			// builtinHistory is ignored when a user is logged in.
// 			builtinHistory: map[string]int{"ignored": 1},
// 			setupUser: func(db *gorm.DB) *user.User {
// 				return &user.User{
// 					Username:   "test",
// 					HistoryMap: map[string]int{"echo": 3, "pwd": 1},
// 				}
// 			},
// 			expectedErr:    nil,
// 			expectedOutput: "echo: 3\npwd: 1\n",
// 			assertHistory:  func(t *testing.T, cmd *HistoryCommand, db *gorm.DB) {},
// 		},
// 		{
// 			name:           "Display user history - empty",
// 			args:           []string{},
// 			builtinHistory: map[string]int{"ignored": 1},
// 			setupUser: func(db *gorm.DB) *user.User {
// 				return &user.User{
// 					Username:   "test",
// 					HistoryMap: map[string]int{},
// 				}
// 			},
// 			expectedErr:    utils.ErrEmptyHistory,
// 			expectedOutput: "",
// 			assertHistory:  func(t *testing.T, cmd *HistoryCommand, db *gorm.DB) {},
// 		},
// 		{
// 			name:           "Invalid args - more than one argument",
// 			args:           []string{"arg1", "arg2"},
// 			builtinHistory: map[string]int{"ls": 2},
// 			setupUser: func(db *gorm.DB) *user.User {
// 				return &user.User{}
// 			},
// 			expectedErr:    utils.ErrInvalidArgs,
// 			expectedOutput: "",
// 			assertHistory:  func(t *testing.T, cmd *HistoryCommand, db *gorm.DB) {},
// 		},
// 		{
// 			name:           "Clean builtin history",
// 			args:           []string{"clean"},
// 			builtinHistory: map[string]int{"ls": 2, "pwd": 1},
// 			setupUser: func(db *gorm.DB) *user.User {
// 				return &user.User{} // No logged-in user
// 			},
// 			expectedErr:    nil,
// 			expectedOutput: "",
// 			assertHistory: func(t *testing.T, cmd *HistoryCommand, db *gorm.DB) {
// 				if len(*cmd.BuiltinHistory) != 0 {
// 					t.Errorf("expected builtin history to be empty after clean, got %v", *cmd.BuiltinHistory)
// 				}
// 			},
// 		},
// 		{
// 			name:           "Clean user history",
// 			args:           []string{"clean"},
// 			builtinHistory: map[string]int{"ignored": 1},
// 			setupUser: func(db *gorm.DB) *user.User {
// 				return &user.User{
// 					Username:   "test",
// 					HistoryMap: map[string]int{"echo": 3, "pwd": 1},
// 				}
// 			},
// 			expectedErr:    nil,
// 			expectedOutput: "",
// 			assertHistory: func(t *testing.T, cmd *HistoryCommand, db *gorm.DB) {
// 				if len(cmd.User.HistoryMap) != 0 {
// 					t.Errorf("expected user history to be empty after clean, got %v", cmd.User.HistoryMap)
// 				}
// 			},
// 		},
// 	}

// 	// Override user.Update for testing clean command on a logged-in user.
// 	// Save the original function to restore later.
// 	origUpdate := user.Update
// 	user.Update = func(db *gorm.DB, u *user.User) error {
// 		// Simulate successful update.
// 		return nil
// 	}
// 	defer func() {
// 		user.Update = origUpdate
// 	}()

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Using a dummy gorm.DB instance (won't be actually used in these tests).
// 			db := &gorm.DB{}
// 			userObj := tt.setupUser(db)
// 			cmd := NewHistoryCommand(&tt.builtinHistory, userObj, db)
// 			var buf bytes.Buffer

// 			err := cmd.Execute(tt.args, &buf)
// 			if err != tt.expectedErr {
// 				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
// 			}

// 			if buf.String() != tt.expectedOutput {
// 				t.Errorf("expected output %q, got %q", tt.expectedOutput, buf.String())
// 			}

// 			if tt.assertHistory != nil {
// 				tt.assertHistory(t, cmd, db)
// 			}
// 		})
// 	}
// }
