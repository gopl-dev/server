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
	"token":    z.String().Required(z.Message("Token is required")),
	"password": newPasswordInputRules,
}

var (
	// ErrInvalidPasswordResetToken ...
	ErrInvalidPasswordResetToken = app.ErrUnprocessable("password reset request is either expired or invalid")
)

// ResetPassword handles the logic for resetting a user's password using a token.
// It validates the token, checks for expiration, updates the user's password, and deletes the token.
func (s *Service) ResetPassword(ctx context.Context, token, password string) error {
	ctx, span := s.tracer.Start(ctx, "ResetPassword")
	defer span.End()

	err := ValidateResetPasswordInput(&token, &password)
	if err != nil {
		return err
	}

	prt, err := s.db.FindPasswordResetToken(ctx, token)
	if errors.Is(err, repo.ErrPasswordResetTokenNotFound) {
		return ErrInvalidPasswordResetToken
	}
	if err != nil {
		return err
	}
	if prt.Invalid() {
		return ErrInvalidPasswordResetToken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), app.DefaultBCryptCost)
	if err != nil {
		return err
	}

	err = s.db.UpdateUserPassword(ctx, prt.UserID, string(passwordHash))
	if err != nil {
		return err
	}

	return s.db.DeletePasswordResetToken(ctx, prt.ID)
}

// ValidateResetPasswordInput ...
func ValidateResetPasswordInput(token, password *string) (err error) {
	in := &ResetPasswordInput{
		Token:    *token,
		Password: *password,
	}

	in.Token = strings.TrimSpace(in.Token)
	in.Password = strings.TrimSpace(in.Password)

	err = validateInput(resetPasswordInputRules, in)
	if err != nil {
		return
	}

	*token = in.Token
	*password = in.Password
	return nil
}

// ResetPasswordInput ...
type ResetPasswordInput struct {
	Token    string
	Password string
}
