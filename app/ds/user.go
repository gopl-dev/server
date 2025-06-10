package ds

import (
	"regexp"
	"time"

	z "github.com/Oudwins/zog"
)

var usernameBasicRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
var usernameSpecialCharsRegex = regexp.MustCompile(`^[^._-]*([._-][^._-]*){0,2}$`)

var UserValidationRules = z.Shape{
	"username": z.String().Min(2).Max(30).Required(z.Message("Username is required")).
		Match(usernameBasicRegex, z.Message("Username can only contain letters, numbers, dots, underscores, and dashes")).
		Match(usernameSpecialCharsRegex, z.Message("Username cannot contain more than two dots, underscores, or dashes")),
	"email":    z.String().Email().Required(z.Message("Email is required")),
	"password": z.String().Min(6).Required(z.Message("Password is required")),
}

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
