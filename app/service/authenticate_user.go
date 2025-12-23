package service

import (
	"context"
	"errors"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"golang.org/x/crypto/bcrypt"
)

var authenticateUserInputRules = z.Shape{
	"Email":    emailInputRules,
	"Password": z.String().Required(z.Message("Password is required")),
}

var (
	// ErrInvalidEmailOrPassword is returned when a user attempts to log in with credentials
	// that do not match any record.
	ErrInvalidEmailOrPassword = app.ErrUnprocessable("invalid email or password")
)

// AuthenticateUser authenticates a user using their email and password.
func (s *Service) AuthenticateUser(ctx context.Context, email, password string) (
	user *ds.User, token string, err error) {
	ctx, span := s.tracer.Start(ctx, "AuthenticateUser")
	defer span.End()

	err = ValidateAuthenticateUserInput(&email, &password)
	if err != nil {
		return
	}

	user, err = s.db.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			err = ErrInvalidEmailOrPassword
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		err = ErrInvalidEmailOrPassword
		return
	}

	sess, err := s.CreateUserSession(ctx, user.ID)
	if err != nil {
		return
	}

	token, err = app.NewSignedSessionJWT(sess.ID.String(), user.ID)
	return
}

// AuthenticateUserInput ...
type AuthenticateUserInput struct {
	Email, Password string
}

// ValidateAuthenticateUserInput ...
func ValidateAuthenticateUserInput(email, password *string) (err error) {
	in := &AuthenticateUserInput{
		Email:    *email,
		Password: *password,
	}

	in.Email = strings.TrimSpace(in.Email)
	in.Password = strings.TrimSpace(in.Password)

	err = validateInput(authenticateUserInputRules, in)
	if err != nil {
		return
	}

	*email, *password = in.Email, in.Password
	return nil
}
