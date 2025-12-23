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

var createChangeEmailRequestInputRules = z.Shape{
	"UserID":   userIDInputRules,
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
func (s *Service) CreateChangeEmailRequest(ctx context.Context, userID int64, newEmail string) (err error) {
	ctx, span := s.tracer.Start(ctx, "CreateChangeEmailRequest")
	defer span.End()

	err = ValidateCreateChangeEmailRequestInput(userID, &newEmail)
	if err != nil {
		return
	}

	user, err := s.db.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Email == newEmail {
		return ErrChangeEmailToSameEmail
	}

	// Check if the new email is already taken by another user.
	existingUser, err := s.db.FindUserByEmail(ctx, newEmail)
	if errors.Is(err, repo.ErrUserNotFound) {
		err = nil
	}
	if err != nil {
		return err
	}
	if existingUser != nil && existingUser.ID != user.ID {
		return app.InputError{"email": UserWithThisEmailAlreadyExists}
	}

	token, err := app.Token(emailChangeTokenLength)
	if err != nil {
		return err
	}

	req := &ds.ChangeEmailRequest{
		UserID:    user.ID,
		NewEmail:  newEmail,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour * 1),
		CreatedAt: time.Now(),
	}

	err = s.db.CreateChangeEmailRequest(ctx, req)
	if err != nil {
		return err
	}

	// TODO: Send email asynchronously
	return email.Send(newEmail, email.ConfirmEmailChange{
		Username: user.Username,
		Token:    token,
	})
}

// ValidateCreateChangeEmailRequestInput ...
func ValidateCreateChangeEmailRequestInput(userID int64, newEmail *string) (err error) {
	in := &CreateChangeEmailRequestInput{
		UserID:   userID,
		NewEmail: *newEmail,
	}

	in.NewEmail = strings.TrimSpace(in.NewEmail)

	err = validateInput(createChangeEmailRequestInputRules, in)
	if err != nil {
		return
	}

	*newEmail = in.NewEmail
	return nil
}

// CreateChangeEmailRequestInput ...
type CreateChangeEmailRequestInput struct {
	UserID   int64
	NewEmail string
}
