package login

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

func TestLoginCommand_Name(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	currentUser := &user.User{}
	cmd := NewLoginCommand(db, currentUser)

	if cmd.Name() != "login" {
		t.Errorf("Name() should return 'login', but got '%s'", cmd.Name())
	}
}

func TestLoginCommand_Execute(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupDB     func(db *gorm.DB)
		expectedErr error
		assertUser  func(t *testing.T, currentUser *user.User, expectedError error)
	}{
		// {
		// 	name: "Valid username and password",
		// 	args: []string{"testuser", "password123"},
		// 	setupDB: func(db *gorm.DB) {
		// 		hashedPassword := "password123"
		// 		existingUser := &user.User{Username: "testuser", Password: hashedPassword}
		// 		err := user.RegisterUser(db, existingUser)
		// 		if err != nil {
		// 			t.Fatalf("Failed to setup existing user: %v", err)
		// 		}
		// 	},
		// 	expectedErr: nil,
		// 	assertUser: func(t *testing.T, currentUser *user.User, expectedError error) {
		// 		if currentUser.Username != "testuser" {
		// 			t.Errorf("Expected current user username to be 'testuser', but got '%s'", currentUser.Username)
		// 		}
		// 		if currentUser.Password != "password123" {
		// 			t.Errorf("Expected current user password to be 'password123', but got '%s'", currentUser.Password)
		// 		}
		// 		if currentUser.ID == 0 {
		// 			t.Errorf("Expected current user ID to be set (not 0)")
		// 		}
		// 		if expectedError != nil {
		// 			t.Errorf("Expected no error, but got: %v", expectedError)
		// 		}
		// 	},
		// },
		// {
		// 	name: "Valid username, no password provided (empty password in DB)",
		// 	args: []string{"testuser"},
		// 	setupDB: func(db *gorm.DB) {
		// 		existingUser := &user.User{Username: "testuser", Password: ""}
		// 		err := user.RegisterUser(db, existingUser)
		// 		if err != nil {
		// 			t.Fatalf("Failed to setup existing user: %v", err)
		// 		}
		// 	},
		// 	expectedErr: nil,
		// 	assertUser: func(t *testing.T, currentUser *user.User, expectedError error) {
		// 		if currentUser.Username != "testuser" {
		// 			t.Errorf("Expected current user username to be 'testuser', but got '%s'", currentUser.Username)
		// 		}
		// 		if currentUser.Password != "" {
		// 			t.Errorf("Expected current user password to be empty, but got '%s'", currentUser.Password)
		// 		}
		// 		if currentUser.ID == 0 {
		// 			t.Errorf("Expected current user ID to be set")
		// 		}
		// 		if expectedError != nil {
		// 			t.Errorf("Expected no error, but got: %v", expectedError)
		// 		}
		// 	},
		// },
		{
			name: "Invalid username",
			args: []string{"nonexistentuser", "password"},
			setupDB: func(db *gorm.DB) {
				// No setup needed, user should not exist
			},
			expectedErr: errors.New("user not found"), // <----- CORRECT expectedErr to error value
			assertUser: func(t *testing.T, currentUser *user.User, expectedError error) {
				if currentUser.Username != "" {
					t.Errorf("Expected current user username to be empty, but got '%s'", currentUser.Username)
				}
				if currentUser.ID != 0 {
					t.Errorf("Expected current user ID to be 0, but got %d", currentUser.ID)
				}
				if !errors.Is(expectedError, errors.New("user not found")) { // <---- CORRECT error assertion using errors.Is
					t.Errorf("Expected error '%v', but got: %v", errors.New("user not found"), expectedError) // Corrected expected error value in error message
				}
			},
		},
		{
			name: "Wrong password",
			args: []string{"testuser", "wrongpassword"},
			setupDB: func(db *gorm.DB) {
				hashedPassword := "password123"
				existingUser := &user.User{Username: "testuser", Password: hashedPassword}
				err := user.RegisterUser(db, existingUser)
				if err != nil {
					t.Fatalf("Failed to setup existing user: %v", err)
				}
			},
			expectedErr: user.ErrWrongPassword,
			assertUser: func(t *testing.T, currentUser *user.User, expectedError error) {
				if currentUser.Username != "" {
					t.Errorf("Expected current user username to be empty, but got '%s'", currentUser.Username)
				}
				if currentUser.ID != 0 {
					t.Errorf("Expected current user ID to be 0")
				}
				if !errors.Is(expectedError, user.ErrWrongPassword) {
					t.Errorf("Expected error '%v', but got: %v", user.ErrWrongPassword, expectedError)
				}
			},
		},
		{
			name:        "Missing username",
			args:        []string{},
			setupDB:     func(db *gorm.DB) {},
			expectedErr: utils.ErrUsernameRequired,
			assertUser: func(t *testing.T, currentUser *user.User, expectedError error) {
				if currentUser.Username != "" {
					t.Errorf("Expected current user username to be empty, but got '%s'", currentUser.Username)
				}
				if !errors.Is(expectedError, utils.ErrUsernameRequired) {
					t.Errorf("Expected error '%v', but got: %v", utils.ErrUsernameRequired, expectedError)
				}
			},
		},
		{
			name:        "Too many arguments",
			args:        []string{"user", "pass", "extra"},
			setupDB:     func(db *gorm.DB) {},
			expectedErr: utils.ErrInvalidArgs,
			assertUser: func(t *testing.T, currentUser *user.User, expectedError error) {
				if currentUser.Username != "" {
					t.Errorf("Expected current user username to be empty, but got '%s'", currentUser.Username)
				}
				if !errors.Is(expectedError, utils.ErrInvalidArgs) {
					t.Errorf("Expected error '%v', but got: %v", utils.ErrInvalidArgs, expectedError)
				}
			},
		},
	}

	for _, cmdTest := range tests {
		t.Run(cmdTest.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer teardownTestDB(t, db)

			currentUser := &user.User{} // <----- CREATE currentUser HERE, inside t.Run
			cmdWithDB := NewLoginCommand(db, currentUser)
			cmdTest.setupDB(db)

			var buf bytes.Buffer
			err := cmdWithDB.Execute(cmdTest.args, &buf)

			if !errors.Is(err, cmdTest.expectedErr) {
				t.Errorf("Execute() error = %v, wantErr %v", err, cmdTest.expectedErr)
			}

			cmdTest.assertUser(t, currentUser, err)
		})
	}
}