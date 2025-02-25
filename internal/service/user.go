package user

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	ErrNoUserFound      = errors.New("no user found")
	ErrUserShouldntNill = errors.New("user cannot be nil")
	ErrUserNameRequired = errors.New("username is required")
	ErrWrongPassword    = errors.New("wrong password")
	ErrDuplicateUser    = errors.New("duplicate user exists with this username")
)

// ... other code ...

// RegisterUser inserts a new user into the database
func RegisterUser(db *gorm.DB, user *User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	if err := validate(user); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Check if the user already exists
	var existingUser User
	if err := db.Where("user_name = ?", user.Username).First(&existingUser).Error; err == nil {
		return ErrDuplicateUser
	}

	// Initialize History as an empty map[string]int
	historyMap := map[string]int{} // Or make(map[string]int)

	// JSON encode the map
	historyJSON, err := json.Marshal(historyMap)
	if err != nil {
		return fmt.Errorf("failed to encode history to JSON: %w", err)
	}
	user.History = string(historyJSON) // Store JSON string in History field

	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to insert user into database: %w", err)
	}

	return nil
}

// ... other code ...

// GetUser retrieves the user and decodes the command history
func GetUser(db *gorm.DB, username string, password string) (User, error) {
	var user User
	if err := db.Where("user_name = ? ", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user, fmt.Errorf("user not found")
		}
		return user, err
	}
	if password != "" && user.Password != password {
		return user, ErrWrongPassword
	}

	// JSON decode History string back to map[string]int
	var historyMap map[string]int
	err := json.Unmarshal([]byte(user.History), &historyMap)
	if err != nil {
		// Handle error if JSON decoding fails (e.g., data corruption in DB)
		// You might choose to return an error or just return an empty map in case of decoding issues.
		// For now, let's proceed with an empty map if decoding fails.
		historyMap = map[string]int{} // Initialize to empty map if decoding fails
		user.HistoryMap = historyMap
		return user, err
	}
	user.HistoryMap = historyMap

	return user, nil
}

// ... other code ...
func Update(db *gorm.DB, user *User) (err error) {
	if user == nil {
		return ErrUserShouldntNill
	}
	if err := validate(user); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	historyJSON, err := json.Marshal(user.HistoryMap)
	if err != nil {
		return fmt.Errorf("failed to encode history to JSON: %w", err)
	}
	user.History = string(historyJSON)
	err = db.Save(user).Error
	if err != nil {
		return fmt.Errorf("failed to update user in database: %w", err)
	}

	return nil
}

func validate(user *User) (err error) {
	if user.Username == "" {
		return ErrUserNameRequired
	}
	return nil
}
