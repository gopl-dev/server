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

	err = ValidateConfirmEmailInput(&code)
	if err != nil {
		return
	}

	ec, err := s.db.FindEmailConfirmationByCode(ctx, code)
	if err != nil {
		return
	}

	if ec == nil || ec.Invalid() {
		err = app.InputError{"code": InvalidConfirmationCode}

		return
	}

	err = s.db.SetUserEmailConfirmed(ctx, ec.UserID)
	if err != nil {
		return
	}

	err = s.db.DeleteEmailConfirmation(ctx, ec.ID)
	if err != nil {
		return
	}

	return s.LogEmailConfirmed(ctx, ec.UserID)
}

// ValidateConfirmEmailInput ...
func ValidateConfirmEmailInput(code *string) (err error) {
	in := &ConfirmEmailInput{
		Code: *code,
	}

	in.Code = strings.TrimSpace(in.Code)

	err = validateInput(confirmEmailInputRules, in)
	if err != nil {
		return
	}

	*code = in.Code
	return nil
}

// ConfirmEmailInput ...
type ConfirmEmailInput struct {
	Code string
}
