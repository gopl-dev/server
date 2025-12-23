package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/pkg/email"
	"golang.org/x/crypto/bcrypt"
)

// UsernameBasicRegex defines the basic character set allowed in a username (letters, numbers, dot, underscore, dash).
var UsernameBasicRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// UsernameSpecialCharsRegex enforces a limit on the maximum number of special characters (dot, underscore, dash).
var UsernameSpecialCharsRegex = regexp.MustCompile(`^[^._-]*([._-][^._-]*){0,2}$`)

var registerUserInputRules = z.Shape{
	"Username": z.String().Required().
		Min(UsernameMinLen, z.Message("Username must be at least 2 characters")).
		Max(UsernameMaxLen, z.Message("Username must be at most 30 characters")).
		Required(z.Message("Username is required")).
		Match(UsernameBasicRegex,
			z.Message("Username can only contain letters, numbers, dots, underscores, and dashes")).
		Match(UsernameSpecialCharsRegex,
			z.Message("Username cannot contain more than two dots, underscores, or dashes")),
	"Email":    emailInputRules,
	"Password": newPasswordInputRules,
}

const (
	// UserWithThisEmailAlreadyExists is the specific error message for email validation failure during registration.
	UserWithThisEmailAlreadyExists = "User with this email already exists."

	// UsernameAlreadyTaken is the specific error message for username validation failure during registration.
	UsernameAlreadyTaken = "Username already taken"
)

// RegisterUser handles the complete user registration process.
func (s *Service) RegisterUser(ctx context.Context, username, emailAddr, password string) (user *ds.User, err error) {
	ctx, span := s.tracer.Start(ctx, "RegisterUserArgs")
	defer span.End()

	err = ValidateRegisterUserInput(&username, &emailAddr, &password)
	if err != nil {
		return
	}

	_, err = s.db.FindUserByEmail(ctx, emailAddr)
	if err == nil {
		return nil, app.InputError{"email": UserWithThisEmailAlreadyExists}
	}
	if !errors.Is(err, repo.ErrUserNotFound) {
		return nil, err
	}

	_, err = s.db.FindUserByUsername(ctx, username)
	if err == nil {
		return nil, app.InputError{"username": UsernameAlreadyTaken}
	}
	if !errors.Is(err, repo.ErrUserNotFound) {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), app.DefaultBCryptCost)
	if err != nil {
		return
	}

	user = &ds.User{
		Username:       username,
		Email:          emailAddr,
		Password:       string(passwordHash),
		EmailConfirmed: false,
		CreatedAt:      time.Now(),
	}

	err = s.db.CreateUser(ctx, user)
	if err != nil {
		return
	}

	emailConfirmCode, err := s.CreateEmailConfirmation(ctx, user.ID)
	if err != nil {
		return
	}

	// todo send email async
	err = email.Send(user.Email, email.ConfirmEmail{
		Username: user.Username,
		Email:    emailAddr,
		Code:     emailConfirmCode,
	})
	if err != nil {
		return
	}

	err = s.LogUserRegistered(ctx, user.ID)
	return
}

// ValidateRegisterUserInput ...
func ValidateRegisterUserInput(username, email, password *string) (err error) {
	in := &RegisterUserInput{
		Username: *username,
		Email:    *email,
		Password: *password,
	}

	in.Username = strings.TrimSpace(in.Username)
	in.Email = strings.TrimSpace(in.Email)
	in.Password = strings.TrimSpace(in.Password)

	err = validateInput(registerUserInputRules, in)
	if err != nil {
		return
	}

	*username = in.Username
	*email = in.Email
	*password = in.Password
	return nil
}

// RegisterUserInput defines the expected input parameters for the user registration process.
type RegisterUserInput struct {
	Username string
	Email    string
	Password string
}
