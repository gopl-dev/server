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

	in := &AuthenticateUserInput{
		Email:    email,
		Password: password,
	}
	err = Normalize(in)
	if err != nil {
		return
	}

	user, err  = s.db.FindUserByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			err = ErrInvalidEmailOrPassword
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password))
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

// Sanitize ...
func (in *AuthenticateUserInput) Sanitize() {
	in.Email = strings.TrimSpace(in.Email)
	in.Password = strings.TrimSpace(in.Password)
}

// Validate ...
func (in *AuthenticateUserInput) Validate() error {
	return validateInput(authenticateUserInputRules, in)
}
