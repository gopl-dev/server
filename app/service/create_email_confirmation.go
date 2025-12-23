package service

import (
	"context"
	"math/rand"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
)

var createEmailConfirmationInputRules = z.Shape{
	"UserID": userIDInputRules,
}

const (
	emailConfirmationTTL     = time.Hour * 24
	emailConfirmationCodeLen = 6
)

// CreateEmailConfirmation generates a unique confirmation code, calculates its expiration time,
// and saves the email confirmation record to the database for the given user ID.
func (s *Service) CreateEmailConfirmation(ctx context.Context, userID int64) (code string, err error) {
	ctx, span := s.tracer.Start(ctx, "CreateEmailConfirmation")
	defer span.End()

	err = ValidateCreateEmailConfirmationInput(userID)
	if err != nil {
		return
	}

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

// ValidateCreateEmailConfirmationInput ...
func ValidateCreateEmailConfirmationInput(userID int64) (err error) {
	in := &CreateEmailConfirmationInput{
		UserID: userID,
	}

	return validateInput(createEmailConfirmationInputRules, in)
}

// CreateEmailConfirmationInput ...
type CreateEmailConfirmationInput struct {
	UserID int64
}
