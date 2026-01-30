package service

import (
	"context"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
)

var confirmEmailInputRules = z.Shape{
	"code": z.String().Required(z.Message("Code is required")),
}

const (
	// InvalidConfirmationCode is the specific error message returned
	// when an email confirmation code is invalid or expired.
	InvalidConfirmationCode = "Invalid confirmation code"
)

// ConfirmEmail confirms an email address by validating the provided code,
// setting the email_confirmed flag for the associated user, and then deleting the used confirmation record.
func (s *Service) ConfirmEmail(ctx context.Context, code string) (err error) {
	ctx, span := s.tracer.Start(ctx, "ConfirmEmail")
	defer span.End()

	in := &ConfirmEmailInput{Code: code}
	err = Normalize(in)
	if err != nil {
		return
	}

	ec, err := s.db.FindEmailConfirmationByCode(ctx, in.Code)
	if err != nil {
		return
	}

	if ec == nil || ec.Invalid() {
		return app.InputError{"code": InvalidConfirmationCode}
	}

	err = s.db.SetUserEmailConfirmed(ctx, ec.UserID)
	if err != nil {
		return
	}

	err = s.db.DeleteEmailConfirmation(ctx, ec.ID)
	if err != nil {
		return
	}

	user, err := s.FindUserByID(ctx, ec.UserID)
	if err != nil {
		return
	}

	return s.LogEmailConfirmed(ctx, user.Email, user.ID)
}

// ConfirmEmailInput ...
type ConfirmEmailInput struct {
	Code string
}

// Sanitize ...
func (in *ConfirmEmailInput) Sanitize() {
	in.Code = strings.TrimSpace(in.Code)
}

// Validate ...
func (in *ConfirmEmailInput) Validate() error {
	return validateInput(confirmEmailInputRules, in)
}
