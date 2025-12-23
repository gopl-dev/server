package service

import (
	"context"
	"errors"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/pkg/email"
)

var createPasswordResetRequestInputRules = z.Shape{
	"email": emailInputRules,
}

var (
	passwordResetTokenLength = 32
)

// CreatePasswordResetRequest handles the logic for initiating a password reset.
// It finds the user by email, generates a unique token, and sends it to the user's email.
func (s *Service) CreatePasswordResetRequest(ctx context.Context, emailAddr string) (err error) {
	ctx, span := s.tracer.Start(ctx, "CreatePasswordResetRequest")
	defer span.End()

	err = ValidateCreatePasswordResetRequestInput(&emailAddr)
	if err != nil {
		return
	}

	user, err := s.db.FindUserByEmail(ctx, emailAddr)
	if err != nil {
		// If the user is not found, we don't return an error to prevent email enumeration attacks.
		if errors.Is(err, repo.ErrUserNotFound) {
			return nil
		}

		return err
	}

	resetToken, err := app.Token(passwordResetTokenLength)
	if err != nil {
		return err
	}

	token := &ds.PasswordResetToken{
		UserID:    user.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(time.Hour * 1),
		CreatedAt: time.Now(),
	}

	err = s.db.CreatePasswordResetToken(ctx, token)
	if err != nil {
		return err
	}

	// TODO: Send email asynchronously
	return email.Send(user.Email, email.PasswordResetRequest{
		Username: user.Username,
		Token:    resetToken,
	})
}

// ValidateCreatePasswordResetRequestInput ...
func ValidateCreatePasswordResetRequestInput(email *string) (err error) {
	in := &CreatePasswordResetRequestInput{
		Email: *email,
	}

	in.Email = strings.TrimSpace(in.Email)

	err = validateInput(createPasswordResetRequestInputRules, in)
	if err != nil {
		return
	}

	*email = in.Email
	return nil
}

// CreatePasswordResetRequestInput ...
type CreatePasswordResetRequestInput struct {
	Email string
}
