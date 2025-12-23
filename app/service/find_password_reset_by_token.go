package service

import (
	"context"
	"errors"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
)

var findPasswordResetByTokenInputRules = z.Shape{
	"token": z.String().Required(z.Message("Token is required")),
}

// FindPasswordResetByToken ...
func (s *Service) FindPasswordResetByToken(ctx context.Context, token string) (t *ds.PasswordResetToken, err error) {
	ctx, span := s.tracer.Start(ctx, "FindPasswordResetByToken")
	defer span.End()

	err = ValidateFindPasswordResetByTokenInput(token)
	if err != nil {
		return
	}

	t, err = s.db.FindPasswordResetToken(ctx, token)
	if errors.Is(err, repo.ErrPasswordResetTokenNotFound) {
		return nil, ErrInvalidPasswordResetToken
	}
	if err != nil {
		return nil, err
	}

	if t.Invalid() {
		return nil, ErrInvalidPasswordResetToken
	}

	return t, nil
}

// ValidateFindPasswordResetByTokenInput ...
func ValidateFindPasswordResetByTokenInput(token string) (err error) {
	in := &FindPasswordResetByTokenInput{
		Token: token,
	}

	return validateInput(findPasswordResetByTokenInputRules, in)
}

// FindPasswordResetByTokenInput ...
type FindPasswordResetByTokenInput struct {
	Token string
}
