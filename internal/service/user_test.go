package user


import (
	db "asa/shell/internal/database"
	"testing"
	"time"

	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func TestRegisterUser(t *testing.T) {
	db := db.GetDB()

	uniqueUsernamePrefix := "testuser_register_" + generateTestSuffix()

	testCases := []struct {
		name    string
		user    *User
		wantErr error
	}{
		{
			name:    "Successful registration",
			user:    &User{Username: uniqueUsernamePrefix + "success", Password: "password123"},
			wantErr: nil,
		},
		{
			name:    "Nil user",
			user:    nil,
			wantErr: ErrUserShouldntNill,
		},
		{
			name:    "Validation error - username required",
			user:    &User{Password: "password123"},
			wantErr: ErrUserNameRequired,
		},
		{
			name:    "Duplicate user",
			user:    &User{Username: uniqueUsernamePrefix + "duplicate", Password: "password123"},
			wantErr: ErrDuplicateUser, 
		},
	}

	duplicateUser := &User{Username: uniqueUsernamePrefix + "duplicate", Password: "password123"}
	if err := RegisterUser(db, duplicateUser); err != nil && !errors.Is(err, ErrDuplicateUser) { 
		t.Fatalf("Setup failed: Could not create duplicate user for testing: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := RegisterUser(db, tc.user)

			if tc.wantErr != nil {
				if err == nil {
					t.Errorf("Test case '%s': Expected error '%v', but got nil", tc.name, tc.wantErr)
				} else if !errors.Is(err, tc.wantErr) && !strings.Contains(err.Error(), tc.wantErr.Error()) {
					t.Errorf("Test case '%s': Error mismatch:\nexpected error: '%v'\ngot:          '%v'", tc.name, tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("Test case '%s': Unexpected error: %v", tc.name, err)
			} else {
				if tc.user != nil && tc.user.Username != "" { 
					defer cleanupUser(db, tc.user.Username)
				}
			}
		})
	}
	cleanupUser(db, duplicateUser.Username)
}

func TestGetUser(t *testing.T) {
	db := db.GetDB()

	uniqueUsernamePrefix := "testuser_get_" + generateTestSuffix()

	existingUserCorrectPass := &User{Username: uniqueUsernamePrefix + "correctpass", Password: "correctpassword"}
	existingUserWrongPass := &User{Username: uniqueUsernamePrefix + "wrongpass", Password: "correctpassword"}
	nonExistingUser := &User{Username: uniqueUsernamePrefix + "nonexistent"}

	if err := RegisterUser(db, existingUserCorrectPass); err != nil {
		t.Fatalf("Setup failed: Could not register user for testing: %v", err)
	}
	defer cleanupUser(db, existingUserCorrectPass.Username) 

	if err := RegisterUser(db, existingUserWrongPass); err != nil {
		t.Fatalf("Setup failed: Could not register user for testing: %v", err)
	}
	defer cleanupUser(db, existingUserWrongPass.Username) 

	testCases := []struct {
		name       string
		username   string
		password   string
		wantErr    error
		expectUser bool
		checkUser  func(user User) bool 
	}{
		{
			name:       "Successful get user - correct password",
			username:   existingUserCorrectPass.Username,
			password:   existingUserCorrectPass.Password,
			wantErr:    nil,
			expectUser: true,
			checkUser: func(user User) bool {
				return user.Username == existingUserCorrectPass.Username
			},
		},
		{
			name:       "Successful get user - no password provided",
			username:   existingUserCorrectPass.Username,
			password:   "", 
			wantErr:    ErrPassRequired,
			expectUser: true,
			checkUser: func(user User) bool {
				return user.Username == existingUserCorrectPass.Username
			},
		},
		{
			name:       "User not found",
			username:   nonExistingUser.Username,
			password:   "password123",
			wantErr:    ErrUserNotFound,
			expectUser: false,
		},
		{
			name:       "Wrong password",
			username:   existingUserWrongPass.Username,
			password:   "wrongpassword",
			wantErr:    ErrWrongPassword,
			expectUser: true,
			checkUser: func(user User) bool { 
				return user.Username == existingUserWrongPass.Username
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := GetUser(db, tc.username, tc.password)

			if tc.wantErr != nil {
				if err == nil {
					t.Errorf("Test case '%s': Expected error '%v', but got nil", tc.name, tc.wantErr)
				} else if !errors.Is(err, tc.wantErr) && !strings.Contains(err.Error(), tc.wantErr.Error()) {
					t.Errorf("Test case '%s': Error mismatch:\nexpected error: '%v'\ngot:          '%v'", tc.name, tc.wantErr, err)
				}
				if tc.expectUser && user.Username == "" && tc.wantErr != ErrUserNotFound { 
					t.Errorf("Test case '%s': Expected User to be returned even with error (not UserNotFound), but got empty User", tc.name)
				}

			} else if err != nil {
				t.Fatalf("Test case '%s': Unexpected error: %v", tc.name, err)
			} else {
				if !tc.expectUser {
					t.Errorf("Test case '%s': Expected no user to be returned, but got: %v", tc.name, user)
				} else if user.Username == "" {
					t.Errorf("Test case '%s': Expected user to be returned, but got empty User", tc.name)
				}
				if tc.checkUser != nil && !tc.checkUser(user) {
					t.Errorf("Test case '%s': User data check failed", tc.name) 
				}
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	db := db.GetDB()
	uniqueUsernamePrefix := "testuser_update_" + generateTestSuffix()

	baseUser := &User{Username: uniqueUsernamePrefix + "baseuser", Password: "initialpassword", HistoryMap: map[string]int{"cmd1": 1}}
	if err := RegisterUser(db, baseUser); err != nil {
		t.Fatalf("Setup failed: Could not register base user for update tests: %v", err)
	}
	defer cleanupUser(db, baseUser.Username) // Cleanup after tests

	userForUpdate, err := GetUser(db, baseUser.Username, "initialpassword")
	if err != nil {
		t.Fatalf("Setup failed: Could not get user for update tests: %v", err)
	}

	testCases := []struct {
		name      string
		user      *User
		wantErr   error
		checkUser func(username string, expectedHistory map[string]int) bool 
	}{
		{
			name:    "Successful update - password and history",
			user:    &User{ID: userForUpdate.ID, Username: userForUpdate.Username, Password: "newpassword", HistoryMap: map[string]int{"cmd1": 5, "cmd2": 1}},
			wantErr: nil,
			checkUser: func(username string, expectedHistory map[string]int) bool {
				updatedUser, err := GetUser(db, username, "newpassword")
				if err != nil {
					t.Fatalf("CheckUser failed to GetUser after update: %v", err)
					return false
				}
				if updatedUser.Password != "newpassword" {
					t.Errorf("CheckUser: Password not updated correctly, got: %s, expected: newpassword", updatedUser.Password)
					return false
				}
				if !historyMapsEqual(updatedUser.HistoryMap, expectedHistory) {
					t.Errorf("CheckUser: HistoryMap not updated correctly, got: %v, expected: %v", updatedUser.HistoryMap, expectedHistory)
					return false
				}
				return true
			},
		},
		{
			name:    "Nil user",
			user:    nil,
			wantErr: ErrUserShouldntNill,
		},
		{
			name:    "Validation error - username required",                   
			user:    &User{ID: userForUpdate.ID, Password: "anotherpassword"}, 
			wantErr: ErrUserNameRequired,
		},
		{
			name:    "Update with empty HistoryMap - should clear history",
			user:    &User{ID: userForUpdate.ID, Username: userForUpdate.Username, Password: "password", HistoryMap: map[string]int{}},
			wantErr: nil,
			checkUser: func(username string, expectedHistory map[string]int) bool {
				updatedUser, err := GetUser(db, username, "password")
				if err != nil {
					t.Fatalf("CheckUser failed to GetUser after update with empty HistoryMap: %v", err)
					return false
				}
				if len(updatedUser.HistoryMap) != 0 {
					t.Errorf("CheckUser: HistoryMap not cleared (empty), got: %v, expected empty map", updatedUser.HistoryMap)
					return false
				}
				return true
			},
		},
		{
			name:    "Update without HistoryMap in struct - should modify history (it's assumption)",
			user:    &User{ID: userForUpdate.ID, Username: userForUpdate.Username, Password: "password_only_update"}, 
			wantErr: nil,
			checkUser: func(username string, expectedHistory map[string]int) bool { 
				updatedUser, err := GetUser(db, username, "password_only_update")
				if err != nil {
					t.Fatalf("CheckUser failed to GetUser after update without HistoryMap: %v", err)
					return false
				}
				if !historyMapsEqual(updatedUser.HistoryMap, map[string]int{}) { 
					t.Errorf("CheckUser: HistoryMap not modified when not expected, got: %v, expected original: %v", updatedUser.HistoryMap, map[string]int{})
					return false
				}
				if updatedUser.Password != "password_only_update" {
					t.Errorf("CheckUser: Password not updated, got: %s, expected: password_only_update", updatedUser.Password)
					return false
				}

				return true
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Update(db, tc.user)

			if tc.wantErr != nil {
				if err == nil {
					t.Errorf("Test case '%s': Expected error '%v', but got nil", tc.name, tc.wantErr)
				} else if !errors.Is(err, tc.wantErr) && !strings.Contains(err.Error(), tc.wantErr.Error()) {
					t.Errorf("Test case '%s': Error mismatch:\nexpected error: '%v'\ngot:          '%v'", tc.name, tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("Test case '%s': Unexpected error: %v", tc.name, err)
			} else {
				if tc.checkUser != nil {
					if !tc.checkUser(tc.user.Username, tc.user.HistoryMap) { 
						t.Errorf("Test case '%s': User data check after update failed (see checkUser logs)", tc.name)
					}
				}
			}
		})
	}
}


func cleanupUser(db *gorm.DB, username string) {
	var user User
	db.Where("user_name = ?", username).Delete(&user) 
}

func generateTestSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()) 
}

func historyMapsEqual(map1, map2 map[string]int) bool {
	if len(map1) != len(map2) {
		return false
	}
	for key, val1 := range map1 {
		val2, ok := map2[key]
		if !ok || val1 != val2 {
			return false
		}
	}
	return true
}
