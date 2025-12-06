package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

const emailConfirmationTTL = time.Hour * 24

func (s *Service) CreateEmailConfirmation(ctx context.Context, userID int64) (code string, err error) {
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

// ConfirmEmail confirms an email address by setting the email_confirmed flag for a user.
func (s *Service) ConfirmEmail(ctx context.Context, code string) (err error) {
	ec, err := s.db.FindEmailConfirmationByCode(ctx, code)
	if err != nil {
		return
	}

	if ec == nil || ec.Invalid() {
		err = app.InputError{"code": "Invalid confirmation code"}
		return
	}

	err = s.SetUserEmailConfirmed(ctx, ec.UserID)
	if err != nil {
		return
	}

	err = s.SetUserEmailConfirmed(ctx, ec.UserID)
	if err != nil {
		return
	}

	err = s.db.DeleteEmailConfirmation(ctx, ec.ID)
	return
}

func (s *Service) newEmailConfirmationCode(ctx context.Context) (string, error) {
	chars := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	length := 5
	newCode := func(length int) string {
		token := make([]byte, length)
		for i := 0; i < length; i++ {
			token[i] = chars[rand.Intn(len(chars))]
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
