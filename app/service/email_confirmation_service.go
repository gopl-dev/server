package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
)

const emailConfirmationTTL = time.Hour * 24

func CreateEmailConfirmation(ctx context.Context, userID int64) (code string, err error) {
	code, err = createCode(ctx)
	if err != nil {
		return
	}

	ec := &ds.EmailConfirmation{
		UserID:    userID,
		Code:      code,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(emailConfirmationTTL),
	}

	err = repo.CreateEmailConfirmation(ctx, ec)
	return
}

// ConfirmEmail confirms an email address by setting the email_confirmed flag for a user.
func ConfirmEmail(ctx context.Context, code string) (err error) {
	ec, err := repo.FindEmailConfirmationByCode(ctx, code)
	if err != nil {
		return
	}

	if ec == nil || ec.Invalid() {
		err = app.InputError{"code": "Invalid confirmation code"}
		return
	}

	err = SetUserEmailConfirmed(ctx, ec.UserID)
	if err != nil {
		return
	}

	err = SetUserEmailConfirmed(ctx, ec.UserID)
	if err != nil {
		return
	}

	err = repo.DeleteEmailConfirmation(ctx, ec.ID)
	return
}

func createCode(ctx context.Context) (string, error) {
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
		ec, err := repo.FindEmailConfirmationByCode(ctx, code)
		if err != nil {
			return "", err
		}
		if ec == nil {
			return code, nil
		}

		length++
	}
}
