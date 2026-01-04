package service

import (
	"context"
	"math/rand"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
)

var createEmailConfirmationInputRules = z.Shape{
	"UserID": ds.IDInputRules,
}

const (
	emailConfirmationTTL     = time.Hour * 24
	emailConfirmationCodeLen = 6
)

// CreateEmailConfirmation generates a unique confirmation code, calculates its expiration time,
// and saves the email confirmation record to the database for the given user ID.
func (s *Service) CreateEmailConfirmation(ctx context.Context, userID ds.ID) (code string, err error) {
	ctx, span := s.tracer.Start(ctx, "CreateEmailConfirmation")
	defer span.End()

	in := &CreateEmailConfirmationInput{UserID: userID}
	err = Normalize(in)
	if err != nil {
		return
	}

	code, err = s.newEmailConfirmationCode(ctx)
	if err != nil {
		return
	}

	ec := &ds.EmailConfirmation{
		UserID:    in.UserID,
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

// CreateEmailConfirmationInput ...
type CreateEmailConfirmationInput struct {
	UserID ds.ID
}

// Sanitize ...
func (in *CreateEmailConfirmationInput) Sanitize() {
}

// Validate ...
func (in *CreateEmailConfirmationInput) Validate() error {
	return validateInput(createEmailConfirmationInputRules, in)
}
