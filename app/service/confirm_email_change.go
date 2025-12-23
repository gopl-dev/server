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

	err = ValidateConfirmEmailChangeInput(&token)
	if err != nil {
		return
	}

	req, err := s.db.FindChangeEmailRequestByToken(ctx, token)
	if errors.Is(err, repo.ErrChangeEmailRequestNotFound) {
		return ErrInvalidChangeEmailToken
	}
	if err != nil {
		return err
	}

	if req.Invalid() {
		return ErrInvalidChangeEmailToken
	}

	err = s.db.UpdateUserEmail(ctx, req.UserID, req.NewEmail)
	if err != nil {
		return err
	}

	return s.db.DeleteChangeEmailRequest(ctx, req.ID)
}

// ValidateConfirmEmailChangeInput ...
func ValidateConfirmEmailChangeInput(token *string) (err error) {
	in := &ConfirmEmailChangeInput{
		Token: *token,
	}

	in.Token = strings.TrimSpace(in.Token)

	err = validateInput(confirmEmailChangeInputRules, in)
	if err != nil {
		return
	}

	*token = in.Token
	return nil
}

// ConfirmEmailChangeInput ...
type ConfirmEmailChangeInput struct {
	Token string
}
