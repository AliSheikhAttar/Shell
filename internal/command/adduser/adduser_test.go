package adduser

import (
	user "asa/shell/internal/service"
	"asa/shell/utils"
	"bytes"
	"errors"
	"os"
	"testing"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&user.User{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func teardownTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, _ := db.DB()
	sqlDB.Close()
	os.Remove("file::memory:?cache=shared")
}

func TestAddUserCommand_Name(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cmd := NewAddUserCommand(db, nil)

	if cmd.Name() != "adduser" {
		t.Errorf("Name() should return 'adduser', but got '%s'", cmd.Name())
	}
}

func TestAddUserCommand_Execute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	cmd := NewAddUserCommand(db, nil)

	tests := []struct {
		name        string
		args        []string
		expectedErr error
		assertUser  func(t *testing.T, db *gorm.DB, username string, expectedError error)
	}{
		{
			name:        "Valid username, no password",
			args:        []string{"testuser"},
			expectedErr: nil,
			assertUser: func(t *testing.T, db *gorm.DB, username string, expectedError error) {
				var u user.User
				if err := db.Where("user_name = ?", username).First(&u).Error; err != nil {
					t.Errorf("Expected user '%s' to be created, but not found in DB: %v", username, err)
				} else if u.Username != username {
					t.Errorf("Retrieved user has incorrect username: got '%s', expected '%s'", u.Username, username)
				}
			},
		},
		{
			name:        "Valid username and password",
			args:        []string{"testuser2", "password123"},
			expectedErr: nil,
			assertUser: func(t *testing.T, db *gorm.DB, username string, expectedError error) {
				var u user.User
				if err := db.Where("user_name = ?", username).First(&u).Error; err != nil {
					t.Errorf("Expected user '%s' to be created, but not found in DB: %v", username, err)
				} else if u.Username != username {
					t.Errorf("Retrieved user has incorrect username: got '%s', expected '%s'", u.Username, username)
				}
			},
		},
		{
			name:        "Missing username",
			args:        []string{},
			expectedErr: utils.ErrUsernameRequired,
			assertUser: func(t *testing.T, db *gorm.DB, username string, expectedError error) {
				var u user.User
				err := db.Where("user_name = ?", username).First(&u).Error
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					t.Errorf("Expected no user to be created, but found user or unexpected error: %v", err)
				}
			},
		},
		{
			name:        "Too many arguments",
			args:        []string{"user", "pass", "extra"},
			expectedErr: utils.ErrInvalidArgs,
			assertUser: func(t *testing.T, db *gorm.DB, username string, expectedError error) {
				var u user.User
				err := db.Where("user_name = ?", username).First(&u).Error
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					t.Errorf("Expected no user to be created, but found user or unexpected error: %v", err)
				}
			},
		},
		{
			name:        "Username already exists",
			args:        []string{"existinguser"},
			expectedErr: user.ErrDuplicateUser,
			assertUser: func(t *testing.T, db *gorm.DB, username string, expectedError error) {
				count := int64(0)
				db.Model(&user.User{}).Where("user_name = ?", username).Count(&count)
				if count != 1 {
					t.Errorf("Expected only 1 user with username '%s' after failed add, but found %d", username, count)
				}
				if !errors.Is(expectedError, user.ErrDuplicateUser) {
					t.Errorf("Expected error '%v' in Execute(), but got '%v'", user.ErrDuplicateUser, expectedError)
				}
			},
		},
	}

	for _, cmdTest := range tests {
		t.Run(cmdTest.name, func(t *testing.T) {
            // SETUP for "Username already exists" test case:
            if cmdTest.name == "Username already exists" {
                existingUser := &user.User{Username: cmdTest.args[0]}
                err := user.RegisterUser(db, existingUser)
                if err != nil {
                    t.Fatalf("Failed to setup existing user for test: %v", err)
                }
            }

			var buf bytes.Buffer
			err := cmd.Execute(cmdTest.args, &buf)

			if !errors.Is(err, cmdTest.expectedErr) {
				t.Errorf("Execute() error = %v, wantErr %v", err, cmdTest.expectedErr)
			}

			usernameToAssert := ""
			if len(cmdTest.args) > 0 {
				usernameToAssert = cmdTest.args[0]
			}
			cmdTest.assertUser(t, db, usernameToAssert, err)
		})
	}
}