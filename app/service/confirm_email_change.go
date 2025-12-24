package service

import (
	"context"
	"errors"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/repo"
)

var confirmEmailChangeInputRules = z.Shape{
	"token": z.String().Required(z.Message("Token is required")),
}

var (
	// ErrInvalidChangeEmailToken ...
	ErrInvalidChangeEmailToken = app.ErrUnprocessable("change email request is expired or invalid")
)

// ConfirmEmailChange handles the logic for finalizing an email change via a token.
func (s *Service) ConfirmEmailChange(ctx context.Context, token string) (err error) {
	ctx, span := s.tracer.Start(ctx, "ConfirmEmailChange")
	defer span.End()

	in := &ConfirmEmailChangeInput{Token: token}
	err = Normalize(in)
	if err != nil {
		return
	}

	req, err := s.db.FindChangeEmailRequestByToken(ctx, in.Token)
	if errors.Is(err, repo.ErrChangeEmailRequestNotFound) {
		return ErrInvalidChangeEmailToken
	}
	if err != nil {
		return
	}

	if req.Invalid() {
		return ErrInvalidChangeEmailToken
	}

	err = s.db.UpdateUserEmail(ctx, req.UserID, req.NewEmail)
	if err != nil {
		return
	}

	return s.db.DeleteChangeEmailRequest(ctx, req.ID)
}

// ConfirmEmailChangeInput ...
type ConfirmEmailChangeInput struct {
	Token string
}

// Sanitize ...
func (in *ConfirmEmailChangeInput) Sanitize() {
	in.Token = strings.TrimSpace(in.Token)
}

// Validate ...
func (in *ConfirmEmailChangeInput) Validate() error {
	return validateInput(confirmEmailChangeInputRules, in)
}
