package exit

import (
	"bytes"
	"errors"
	"testing"

	"gorm.io/gorm"

	userSvc "asa/shell/internal/service"
	"asa/shell/utils"
)

type MockDB struct {
	*gorm.DB
	UpdatedUser *userSvc.User
}

func (m *MockDB) Update(db *gorm.DB, user *userSvc.User) error {
	m.UpdatedUser = user
	return nil
}
func TestExitCommand_Execute_NoMockExit_InvalidArguments(t *testing.T) {
	testCases := []struct {
		name             string
		args             []string
		mockUser         *userSvc.User
		expectedOutput   string
		wantErr          error
		expectUserUpdate bool 
	}{
		{
			name:             "Exit with invalid argument - not a number string",
			args:             []string{"abc"},
			mockUser:         &userSvc.User{Username: ""},
			expectedOutput:   "",
			wantErr:          utils.ErrInvalidArgs,
			expectUserUpdate: false,
		},
		{
			name:             "Exit with invalid argument - floating point number string",
			args:             []string{"1.5"},
			mockUser:         &userSvc.User{Username: ""},
			expectedOutput:   "",
			wantErr:          utils.ErrInvalidArgs,
			expectUserUpdate: false,
		},
		{
			name:             "Exit with invalid argument - empty string",
			args:             []string{""},
			mockUser:         &userSvc.User{Username: ""},
			expectedOutput:   "",
			wantErr:          utils.ErrInvalidArgs,
			expectUserUpdate: false,
		},
		{
			name:             "Exit with invalid argument - whitespace string",
			args:             []string{"   "},
			mockUser:         &userSvc.User{Username: ""},
			expectedOutput:   "",
			wantErr:          utils.ErrInvalidArgs,
			expectUserUpdate: false,
		},
		{
			name:             "Exit with too many arguments - two args",
			args:             []string{"1", "2"},
			mockUser:         &userSvc.User{Username: ""},
			expectedOutput:   "",
			wantErr:          utils.ErrTooManyArgs,
			expectUserUpdate: false,
		},
		{
			name:             "Exit with too many arguments - multiple args",
			args:             []string{"arg1", "arg2", "arg3"},
			mockUser:         &userSvc.User{Username: ""},
			expectedOutput:   "",
			wantErr:          utils.ErrTooManyArgs,
			expectUserUpdate: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := &MockDB{UpdatedUser: nil} // Mock DB if testing user update
			cmd := NewExitCommand(mockDB.DB, tc.mockUser)

			var outBuf bytes.Buffer
			err := cmd.Execute(tc.args, &outBuf)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Test case '%s': Expected error type '%v', but got '%v'", tc.name, tc.wantErr, err)
			}

			actualOutput := outBuf.String()
			if actualOutput != tc.expectedOutput {
				t.Errorf("Test case '%s': Output mismatch:\nexpected:\n'%s'\ngot:\n'%s'", tc.name, tc.expectedOutput, actualOutput)
			}

			if tc.expectUserUpdate {
				if mockDB.UpdatedUser == nil || mockDB.UpdatedUser.Username != tc.mockUser.Username {
					t.Errorf("Test case '%s': Expected user update but user was not updated, or incorrect user updated.", tc.name)
				}
			} else {
				if mockDB.UpdatedUser != nil {
					t.Errorf("Test case '%s': Did not expect user update, but user was updated.", tc.name)
				}
			}

			// For invalid arguments, os.Exit should NOT be called, so we cannot check exit code here in this no-mocking scenario for os.Exit.
		})
	}
}

func TestExitCommand_Name_NoMockExit(t *testing.T) {
	mockUser := &userSvc.User{}
	mockDB := &MockDB{}
	cmd := NewExitCommand(mockDB.DB, mockUser)
	if cmd.Name() != "exit" {
		t.Errorf("Name() should return 'exit', but got '%s'", cmd.Name())
	}
}
