package service

import (
	"context"
	"errors"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
)

var findPasswordResetByTokenInputRules = z.Shape{
	"Token": z.String().Required(z.Message("Token is required")),
}

// FindPasswordResetByToken ...
func (s *Service) FindPasswordResetByToken(ctx context.Context, token string) (prt *ds.PasswordResetToken, err error) {
	ctx, span := s.tracer.Start(ctx, "FindPasswordResetByToken")
	defer span.End()

	in := &FindPasswordResetByTokenInput{Token: token}
	err = Normalize(in)
	if err != nil {
		return
	}

	prt, err = s.db.FindPasswordResetToken(ctx, in.Token)
	if errors.Is(err, repo.ErrPasswordResetTokenNotFound) {
		err = ErrInvalidPasswordResetToken
		return
	}
	if err != nil {
		return
	}

	if prt.Invalid() {
		err = ErrInvalidPasswordResetToken
		return
	}

	return
}

// FindPasswordResetByTokenInput ...
type FindPasswordResetByTokenInput struct {
	Token string
}

// Sanitize ...
func (in *FindPasswordResetByTokenInput) Sanitize() {
	in.Token = strings.TrimSpace(in.Token)
}

// Validate ...
func (in *FindPasswordResetByTokenInput) Validate() error {
	return validateInput(findPasswordResetByTokenInputRules, in)
}
