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

var createChangeEmailRequestInputRules = z.Shape{
	"UserID":   ds.IDInputRules,
	"NewEmail": emailInputRules,
}

var (
	// ErrChangeEmailToSameEmail ...
	ErrChangeEmailToSameEmail = app.ErrUnprocessable("you already use this email, no change needed")
)

var (
	emailChangeTokenLength = 32
)

// CreateChangeEmailRequest handles the business logic for a user initiating an email change.
func (s *Service) CreateChangeEmailRequest(ctx context.Context, userID ds.ID, newEmail string) (err error) {
	ctx, span := s.tracer.Start(ctx, "CreateChangeEmailRequest")
	defer span.End()

	in := &CreateChangeEmailRequestInput{
		UserID:   userID,
		NewEmail: newEmail,
	}
	err = Normalize(in)
	if err != nil {
		return
	}

	user, err := s.db.FindUserByID(ctx, userID)
	if err != nil {
		return
	}

	if user.Email == in.NewEmail {
		return ErrChangeEmailToSameEmail
	}

	// Check if the new email is already taken by another user.
	existingUser, err := s.db.FindUserByEmail(ctx, in.NewEmail)
	if errors.Is(err, repo.ErrUserNotFound) {
		err = nil
	}
	if err != nil {
		return
	}
	if existingUser != nil && existingUser.ID != user.ID {
		return app.InputError{"email": UserWithThisEmailAlreadyExists}
	}

	token, err := app.Token(emailChangeTokenLength)
	if err != nil {
		return
	}

	req := &ds.ChangeEmailRequest{
		UserID:    user.ID,
		NewEmail:  in.NewEmail,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour * 1),
		CreatedAt: time.Now(),
	}

	err = s.db.CreateChangeEmailRequest(ctx, req)
	if err != nil {
		return
	}

	err = s.LogEmailChangeRequested(ctx, user.ID)
	if err != nil {
		return
	}

	// TODO: Send email asynchronously
	return email.Send(in.NewEmail, email.ConfirmEmailChange{
		Username: user.Username,
		Token:    token,
	})
}

// CreateChangeEmailRequestInput ...
type CreateChangeEmailRequestInput struct {
	UserID   ds.ID
	NewEmail string
}

// Sanitize ...
func (in *CreateChangeEmailRequestInput) Sanitize() {
	in.NewEmail = strings.TrimSpace(in.NewEmail)
}

// Validate ...
func (in *CreateChangeEmailRequestInput) Validate() error {
	return validateInput(createChangeEmailRequestInputRules, in)
}
