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
	"github.com/gopl-dev/server/email"
)

var createPasswordResetRequestInputRules = z.Shape{
	"Email": emailInputRules,
}

var (
	passwordResetTokenLength = 32
)

// CreatePasswordResetRequest handles the logic for initiating a password reset.
// It finds the user by email, generates a unique token, and sends it to the user's email.
func (s *Service) CreatePasswordResetRequest(ctx context.Context, emailAddr string) (err error) {
	ctx, span := s.tracer.Start(ctx, "CreatePasswordResetRequest")
	defer span.End()

	in := &CreatePasswordResetRequestInput{Email: emailAddr}
	err = Normalize(in)
	if err != nil {
		return
	}

	user, err := s.db.FindUserByEmail(ctx, in.Email)
	if err != nil {
		// If the user is not found, we don't return an error to prevent email enumeration attacks.
		if errors.Is(err, repo.ErrUserNotFound) {
			err = nil
		}

		return
	}

	resetToken, err := app.Token(passwordResetTokenLength)
	if err != nil {
		return
	}

	token := &ds.PasswordResetToken{
		UserID:    user.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(time.Hour * 1),
		CreatedAt: time.Now(),
	}

	err = s.db.CreatePasswordResetToken(ctx, token)
	if err != nil {
		return
	}

	// TODO: Send email asynchronously
	return email.Send(user.Email, email.PasswordResetRequest{
		Username: user.Username,
		Token:    resetToken,
	})
}

// CreatePasswordResetRequestInput ...
type CreatePasswordResetRequestInput struct {
	Email string
}

// Sanitize ...
func (in *CreatePasswordResetRequestInput) Sanitize() {
	in.Email = strings.TrimSpace(in.Email)
}

// Validate ...
func (in *CreatePasswordResetRequestInput) Validate() error {
	return validateInput(createPasswordResetRequestInputRules, in)
}
