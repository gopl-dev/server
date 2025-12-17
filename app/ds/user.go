package ds

import (
	"context"
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

const (
	userCtxKey ctxKey = "user"
)

var passwordValidation = z.String().Min(UserPasswordMinLen).Required(z.Message("Password is required"))
var emailValidation = z.String().Email().Required(z.Message("Email is required"))

// UserValidationRules specifies the validation rules for the User struct.
var UserValidationRules = z.Shape{
	"Username": z.String().Min(UsernameMinLen).Max(UsernameMaxLen).Required(z.Message("Username is required")).
		Match(UsernameBasicRegex,
			z.Message("Username can only contain letters, numbers, dots, underscores, and dashes")).
		Match(UsernameSpecialCharsRegex,
			z.Message("Username cannot contain more than two dots, underscores, or dashes")),
	"Email":    emailValidation,
	"Password": passwordValidation,
}

// ChangePasswordValidationRules ...
var ChangePasswordValidationRules = z.Shape{
	"OldPassword": z.String().Required(z.Message("Password is required")),
	"NewPassword": passwordValidation,
}

// PasswordResetRequestValidationRules ...
var PasswordResetRequestValidationRules = z.Shape{
	"Email": emailValidation,
}

// PasswordResetValidationRules ...
var PasswordResetValidationRules = z.Shape{
	"Token":    z.String().Required(z.Message("Token is required")),
	"Password": passwordValidation,
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

// ToContext adds the given user object to the provided context.
func (u *User) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

// UserFromContext attempts to retrieve user object from the context.
func UserFromContext(ctx context.Context) *User {
	if v := ctx.Value(userCtxKey); v != nil {
		if user, ok := v.(*User); ok {
			return user
		}
	}

	return nil
}
