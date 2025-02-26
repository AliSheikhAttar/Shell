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
	ErrUserNotFound     = errors.New("user not found")
	ErrPassRequired     = errors.New("password required")
)

func RegisterUser(db *gorm.DB, user *User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	if err := validate(user); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	var existingUser User
	err := db.Where("user_name = ?", user.Username).First(&existingUser).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	} else {
		return ErrDuplicateUser
	}

	historyMap := map[string]int{}

	historyJSON, err := json.Marshal(historyMap)
	if err != nil {
		return fmt.Errorf("failed to encode history to JSON: %w", err)
	}
	user.History = string(historyJSON)

	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to insert user into database: %w", err)
	}

	return nil
}

func GetUser(db *gorm.DB, username string, password string) (User, error) {
	var user User
	if err := db.Where("user_name = ? ", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user, ErrUserNotFound
		}
		return user, err
	}
	if user.Password != "" && user.Password != password {
		if password == "" {
			return user, ErrPassRequired
		}
		return user, ErrWrongPassword
	}

	var historyMap map[string]int
	err := json.Unmarshal([]byte(user.History), &historyMap)
	if err != nil {
		historyMap = map[string]int{}
		user.HistoryMap = historyMap
		return user, err
	}
	user.HistoryMap = historyMap

	return user, nil
}

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
