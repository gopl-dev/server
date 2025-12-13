package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

const (
	emailConfirmationTTL     = time.Hour * 24
	emailConfirmationCodeLen = 6
)

const (
	// InvalidConfirmationCode is the specific error message returned
	// when an email confirmation code is invalid or expired.
	InvalidConfirmationCode = "Invalid confirmation code"
)

// CreateEmailConfirmation generates a unique confirmation code, calculates its expiration time,
// and saves the email confirmation record to the database for the given user ID.
func (s *Service) CreateEmailConfirmation(ctx context.Context, userID int64) (code string, err error) {
	ctx, span := s.tracer.Start(ctx, "CreateEmailConfirmation")
	defer span.End()

	code, err = s.newEmailConfirmationCode(ctx)
	if err != nil {
		return
	}

	ec := &ds.EmailConfirmation{
		UserID:    userID,
		Code:      code,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(emailConfirmationTTL),
	}

	err = s.db.CreateEmailConfirmation(ctx, ec)

	return
}

// ConfirmEmail confirms an email address by validating the provided code,
// setting the email_confirmed flag for the associated user, and then deleting the used confirmation record.
func (s *Service) ConfirmEmail(ctx context.Context, code string) (err error) {
	ctx, span := s.tracer.Start(ctx, "ConfirmEmail")
	defer span.End()

	ec, err := s.db.FindEmailConfirmationByCode(ctx, code)
	if err != nil {
		return
	}

	if ec == nil || ec.Invalid() {
		err = app.InputError{"code": InvalidConfirmationCode}

		return
	}

	err = s.SetUserEmailConfirmed(ctx, ec.UserID)
	if err != nil {
		return
	}

	err = s.db.DeleteEmailConfirmation(ctx, ec.ID)
	if err != nil {
		return
	}

	return s.LogEmailConfirmed(ctx, ec.UserID)
}

func (s *Service) newEmailConfirmationCode(ctx context.Context) (string, error) {
	chars := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	length := emailConfirmationCodeLen
	newCode := func(length int) string {
		token := make([]byte, length)
		for i := range length {
			token[i] = chars[rand.Intn(len(chars))] //nolint:gosec
		}

		return string(token)
	}

	for {
		code := newCode(length)

		ec, err := s.db.FindEmailConfirmationByCode(ctx, code)
		if err != nil {
			return "", err
		}

		if ec == nil {
			return code, nil
		}

		length++
	}
}
