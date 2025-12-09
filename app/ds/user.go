package ds

import (
	"regexp"
	"time"

	z "github.com/Oudwins/zog"
)

// UsernameBasicRegex defines the basic character set allowed in a username (letters, numbers, dot, underscore, dash).
var UsernameBasicRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// UsernameSpecialCharsRegex enforces a limit on the maximum number of special characters (dot, underscore, dash).
var UsernameSpecialCharsRegex = regexp.MustCompile(`^[^._-]*([._-][^._-]*){0,2}$`)

const (
	// UsernameMinLen defines the minimum allowed length, in characters, for a user's username.
	UsernameMinLen = 2

	// UsernameMaxLen defines the maximum allowed length, in characters, for a user's username.
	UsernameMaxLen = 30

	// UserPasswordMinLen defines the minimum allowed length, in characters, for a user's password.
	UserPasswordMinLen = 6
)

// UserValidationRules specifies the validation rules for the User struct.
var UserValidationRules = z.Shape{
	"username": z.String().Min(UsernameMinLen).Max(UsernameMaxLen).Required(z.Message("Username is required")).
		Match(UsernameBasicRegex,
			z.Message("Username can only contain letters, numbers, dots, underscores, and dashes")).
		Match(UsernameSpecialCharsRegex,
			z.Message("Username cannot contain more than two dots, underscores, or dashes")),
	"email":    z.String().Email().Required(z.Message("Email is required")),
	"password": z.String().Min(UserPasswordMinLen).Required(z.Message("Password is required")),
}

// User ...
type User struct {
	ID             int64      `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	EmailConfirmed bool       `json:"-"`
	Password       string     `json:"-"`
	CreatedAt      time.Time  `json:"-"`
	UpdatedAt      *time.Time `json:"-"`
	DeletedAt      *time.Time `json:"-"`
}
