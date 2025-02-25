package user

// todo

import (
	db "asa/shell/internal/database"
	"testing"
)

func TestRegister(t *testing.T) {
	db := db.GetDB()
	tests := []struct {
		name  string
		user  *User
		IsErr bool
	}{
		{
			name:  "normal user",
			user:  &User{Username: "testUser"},
			IsErr: false,
		},
		{
			name:  "empty user",
			user:  &User{},
			IsErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := RegisterUser(db, tt.user)
			if (err != nil) != tt.IsErr {
				t.Fatalf("Failed to get current directory: %v", err)
			}

		})
	}

}

func TestInsert(t *testing.T) {
	db := db.GetDB()
	tests := []struct {
		name     string
		user     *User
		command  string
		expected map[string]int
		IsErr    bool
	}{
		{
			name:     "invalid user insertion1",
			user:     &User{Username: "testUser2"},
			command:  "",
			expected: nil,
			IsErr:    false,
		},
		{
			name:     "normal insertion1",
			user:     &User{Username: "testUser"},
			command:  "cd",
			expected: map[string]int{"cd": 1},
			IsErr:    false,
		},
		{
			name:     "normal insertion2",
			user:     &User{Username: "testUser"},
			command:  "pwd",
			expected: map[string]int{"cd": 1, "pwd": 1},
			IsErr:    false,
		},
		{
			name:     "duplicate command insertion",
			user:     &User{Username: "testUser"},
			command:  "cd",
			expected: map[string]int{"cd": 2, "pwd": 1},
			IsErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := Update(db, tt.user)
			if (err != nil) != tt.IsErr {
				t.Fatalf("Failed to get current directory: %v", err)
			}
			if tt.expected != nil {
				user, err := GetUser(db, tt.user.Username, "")
				if (err != nil) != tt.IsErr {
					t.Fatalf("Failed to get current directory: %v", err)
				}
				commands := user.HistoryMap
				if err != nil {
					t.Fatalf("Failed to get current directory: %v", err)
				}
				if len(commands) != len(tt.expected) {
					t.Fatalf("result not equal: %v", err)
				}
				for key, val1 := range commands {
					val2, ok := tt.expected[key]
					if !ok {
						t.Fatalf("result not equal: %v", err)
					}
					if val1 != val2 {
						t.Fatalf("result not equal: %v", err)
					}
				}
			}

		})

	}

}
