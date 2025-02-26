package logout

import (
	db "asa/shell/internal/database"
	user "asa/shell/internal/service"
	"asa/shell/utils"
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type MockUser struct {
	MockUpdateError error
	UserID          int
	Username        string
	Password        string
}

func TestNewLogoutCommand(t *testing.T) {
	db := db.GetDB()
	cmd := NewLogoutCommand(
		db,
		&user.User{Username: "test"},
	)

	if cmd == nil {
		t.Errorf("NewLogoutCommand should not return nil")
	}

	expectedType := "*logout.LogoutCommand"
	actualType := reflect.TypeOf(cmd).String()
	if actualType != expectedType {
		t.Errorf("NewLogoutCommand returned wrong type: got %v want %v", actualType, expectedType)
	}
	if cmd.user == nil {
		t.Errorf("NewLogoutCommand did not set user correctly")
	}
}

func TestLogoutCommand_Name(t *testing.T) {
	db := db.GetDB()
	cmd := NewLogoutCommand(
		db,
		&user.User{},
	)

	expectedName := "logout"
	actualName := cmd.Name()
	if actualName != expectedName {
		t.Errorf("Name should return '%s', but got '%s'", expectedName, actualName)
	}
}

func TestLogoutCommand_Execute_InvalidArgs(t *testing.T) {
	db := db.GetDB()
	cmd := NewLogoutCommand(
		db,
		&user.User{},
	)

	args := []string{"extra_arg"}
	var stdout bytes.Buffer
	err := cmd.Execute(args, &stdout)

	if !errors.Is(err, utils.ErrInvalidArgs) {
		t.Errorf("Execute with invalid args should return ErrInvalidArgs, but got: %v", err)
	}
}

func TestLogoutCommand_Execute_UpdateError(t *testing.T) {
	db := db.GetDB()
	cmd := NewLogoutCommand(
		db,
		&user.User{Username: "testuser"},
	)

	args := []string{}
	var stdout bytes.Buffer
	err := cmd.Execute(args, &stdout)

	if err == nil {
		t.Errorf("Execute should return error when user.Update fails, but got nil")
	}
	if !strings.Contains(err.Error(), "failed to update user in database: ERROR: duplicate key value violates unique constraint") {
		t.Errorf("Execute should return the error from user.Update, got: %v, want: %v", err, errors.New("failed to update user in database: ERROR: duplicate key value violates unique constraint"))
	}
}

func TestLogoutCommand_Execute_Success(t *testing.T) {
	db := db.GetDB()
	var testUser user.User
	testUser, err := user.GetUser(db, "testuser", "")
	if err == nil {
		testUser.HistoryMap = make(map[string]int)
	} else {
		testUser = user.User{Username: "testuser"}
		user.RegisterUser(db, &testUser)
		testUser, _ = user.GetUser(db, "testuser", "")
	}
	cmd := NewLogoutCommand(
		db,
		&testUser,
	)

	args := []string{}
	var stdout bytes.Buffer
	err = cmd.Execute(args, &stdout)

	if err != nil {
		t.Errorf("Execute should not return error on success, but got: %v", err)
	}

	// Assert user object is reset - you need to access the user from LogoutCommand
	// To do this effectively in test, you might need to make the user field in LogoutCommand accessible for testing,
	// or use a getter if you don't want to export it directly.
	// For this example, let's assume you can access cmd.user directly for testing purposes.
	if cmd.user.ID != 0 {
		t.Errorf("User ID should be reset to 0, but got: %d", cmd.user.ID)
	}
	if cmd.user.Username != "" {
		t.Errorf("Username should be reset to '', but got: '%s'", cmd.user.Username)
	}
	if cmd.user.Password != "" {
		t.Errorf("Password should be reset to '', but got: '%s'", cmd.user.Password)
	}
}

func TestLogoutCommand_Execute_Success_Stdout(t *testing.T) {
	db := db.GetDB()
	var testUser user.User
	testUser, err := user.GetUser(db, "testuser", "")
	if err == nil {
		testUser.HistoryMap = make(map[string]int)
	} else {
		testUser = user.User{Username: "testuser"}
		user.RegisterUser(db, &testUser)
		testUser, _ = user.GetUser(db, "testuser", "")
	}
	cmd := NewLogoutCommand(
		db,
		&testUser,
	)

	args := []string{}
	var stdout bytes.Buffer
	err = cmd.Execute(args, &stdout)

	if err != nil {
		t.Errorf("Execute should not return error on success, but got: %v", err)
	}

	// In this case logout command doesn't write anything to stdout based on code provided.
	expectedStdout := ""
	actualStdout := stdout.String()
	if actualStdout != expectedStdout {
		t.Errorf("Stdout should be '%s', but got '%s'", expectedStdout, actualStdout)
	}
}

func TestLogoutCommand_Execute_Success_Update(t *testing.T) {
	db := db.GetDB()
	var testUser user.User
	testUser, err := user.GetUser(db, "testuser", "")
	if err == nil {
		testUser.HistoryMap = make(map[string]int)
	} else {
		testUser = user.User{Username: "testuser"}
		user.RegisterUser(db, &testUser)
	}
	cmd := NewLogoutCommand(
		db,
		&testUser,
	)
	args := []string{}
	var stdout bytes.Buffer
	testUser.HistoryMap["testcommand"] = 1
	err = cmd.Execute(args, &stdout)

	if err != nil {
		t.Errorf("Execute should not return error on success, but got: %v", err)
	}

	// In this case logout command doesn't write anything to stdout based on code provided.
	expectedHistoryMap := map[string]int{"testcommand": 1}
	testUserAfterLogout, err := user.GetUser(db, "testuser", "")
	if err != nil {
		t.Errorf("user not exist")
	}

	actualHistoryMap := testUserAfterLogout.HistoryMap
	if !equalMaps(actualHistoryMap, expectedHistoryMap) {
		t.Errorf("HistoryMap should be '%v', but got '%v'", expectedHistoryMap, actualHistoryMap)
	}
}

func equalMaps(map1 map[string]int, map2 map[string]int) bool {
	if len(map1) != len(map2) {
		return false
	}
	for key, value := range map1 {
		if val, ok := map2[key]; !ok || value != val {
			return false
		}
	}
	return true
}
