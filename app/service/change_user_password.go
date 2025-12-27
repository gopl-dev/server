package service

import (
	"context"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"golang.org/x/crypto/bcrypt"
)

var changePasswordInputRules = z.Shape{
	"UserID":      idInputRules,
	"OldPassword": z.String().Required(z.Message("Password is required")),
	"NewPassword": newPasswordInputRules,
}

var (
	// ErrInvalidPassword is returned when a user tries to change their password
	// but provides an incorrect old password.
	ErrInvalidPassword = app.ErrUnprocessable("invalid password")
)

// ChangeUserPassword handles the logic for an authenticated user to change their own password.
func (s *Service) ChangeUserPassword(ctx context.Context, userID ds.ID, oldPassword, newPassword string) (err error) {
	ctx, span := s.tracer.Start(ctx, "ChangeUserPassword")
	defer span.End()

	in := &ChangeUserPasswordInput{
		UserID:      userID,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}
	err = Normalize(in)
	if err != nil {
		return err
	}

	user, err := s.FindUserByID(ctx, in.UserID)
	if user == nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.OldPassword))
	if err != nil {
		return app.InputError{"old_password": ErrInvalidPassword.Error()}
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), app.DefaultBCryptCost)
	if err != nil {
		return err
	}

	return s.db.UpdateUserPassword(ctx, user.ID, string(newPasswordHash))
}

// ChangeUserPasswordInput defines the input for changing a user's password.
type ChangeUserPasswordInput struct {
	UserID      ds.ID
	OldPassword string
	NewPassword string
}

// Sanitize ...
func (in *ChangeUserPasswordInput) Sanitize() {
	in.OldPassword = strings.TrimSpace(in.OldPassword)
	in.NewPassword = strings.TrimSpace(in.NewPassword)
}

// Validate ...
func (in *ChangeUserPasswordInput) Validate() error {
	return validateInput(changePasswordInputRules, in)
}
