package ds

import (
	"testing"

	z "github.com/Oudwins/zog"
)

func TestUsernameValidation(t *testing.T) {
	type user struct {
		Username string
		Email    string
		Password string
	}

	tests := []struct {
		name     string
		username string
		valid    bool
	}{
		// Valid usernames
		{"simple username", "johnDoe123", true},
		{"username with one dot", "john.doe", true},
		{"username with two dots", "john.doe.123", true},
		{"username with one underscore", "john_doe", true},
		{"username with two underscores", "john__doe", true},
		{"username with one dash", "john-doe", true},
		{"username with two dashes", "john--doe", true},
		{"username with mixed special chars", "john.doe_123456", true},

		// Invalid usernames
		{"empty username", "", false},
		{"three dots", "john.doe.123.456", false},
		{"username with mixed special chars", "john.doe_123-456", false},
		{"three underscores", "john_doe_doe_doe", false},
		{"three dashes", "john-doe-doe-doe", false},
		{"only special chars", "._-", false},
		{"contains spaces", "john doe", false},
		{"contains other special chars", "john@doe", false},
		{"too short", "j", false},
		{"too long", "johndoejohndoejohndoejohndoejohndoe", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic regex
			if !usernameBasicRegex.MatchString(tt.username) && tt.valid {
				t.Errorf("usernameBasicRegex.MatchString() = false, want true for username: %s", tt.username)
			}

			// Test special chars regex
			if !usernameSpecialCharsRegex.MatchString(tt.username) && tt.valid {
				t.Errorf("usernameSpecialCharsRegex.MatchString() = false, want true for username: %s", tt.username)
			}

			// Test full validation rules
			err := z.Struct(UserValidationRules).Validate(&user{
				Username: tt.username,
				Email:    "test@example.com", // Required for validation
				Password: "password123",      // Required for validation
			})

			if (err != nil) == tt.valid {
				t.Errorf("UserValidationRules.Validate() error = %v, wantErr %v for username: %s", err, tt.valid, tt.username)
			}
		})
	}
}
