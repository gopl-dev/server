package service

import (
	"context"
	"errors"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/repo"
	"golang.org/x/crypto/bcrypt"
)

// PasswordResetValidationRules ...
var resetPasswordInputRules = z.Shape{
	"Token":    z.String().Required(z.Message("Token is required")),
	"Password": newPasswordInputRules,
}

var (
	// ErrInvalidPasswordResetToken ...
	ErrInvalidPasswordResetToken = app.ErrUnprocessable("password reset request is either expired or invalid")
)

// ResetPassword handles the logic for resetting a user's password using a token.
// It validates the token, checks for expiration, updates the user's password, and deletes the token.
func (s *Service) ResetPassword(ctx context.Context, token, password string) (err error) {
	ctx, span := s.tracer.Start(ctx, "ResetPassword")
	defer span.End()

	in := &ResetPasswordInput{
		Token:    token,
		Password: password,
	}
	err = Normalize(in)
	if err != nil {
		return
	}

	prt, err := s.db.FindPasswordResetToken(ctx, in.Token)
	if errors.Is(err, repo.ErrPasswordResetTokenNotFound) {
		err = ErrInvalidPasswordResetToken
		return
	}
	if err != nil {
		return
	}
	if prt.Invalid() {
		err = ErrInvalidPasswordResetToken
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), app.DefaultBCryptCost)
	if err != nil {
		return err
	}

	err = s.db.UpdateUserPassword(ctx, prt.UserID, string(passwordHash))
	if err != nil {
		return err
	}

	return s.db.DeletePasswordResetToken(ctx, prt.ID)
}

// ResetPasswordInput ...
type ResetPasswordInput struct {
	Token    string
	Password string
}

// Sanitize ...
func (in *ResetPasswordInput) Sanitize() {
	in.Token = strings.TrimSpace(in.Token)
	in.Password = strings.TrimSpace(in.Password)
}

// Validate ...
func (in *ResetPasswordInput) Validate() error {
	return validateInput(resetPasswordInputRules, in)
}
