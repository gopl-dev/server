package service

import (
	"context"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
	"golang.org/x/crypto/bcrypt"
)

var changePasswordInputRules = z.Shape{
	"UserID":      userIDInputRules,
	"OldPassword": z.String().Required(z.Message("Password is required")),
	"NewPassword": newPasswordInputRules,
}

var (
	// ErrInvalidPassword is returned when a user tries to change their password
	// but provides an incorrect old password.
	ErrInvalidPassword = app.ErrUnprocessable("invalid password")
)

// ChangeUserPassword handles the logic for an authenticated user to change their own password.
func (s *Service) ChangeUserPassword(ctx context.Context, userID int64, oldPassword, newPassword string) (err error) {
	ctx, span := s.tracer.Start(ctx, "ChangeUserPassword")
	defer span.End()

	err = ValidateChangeUserPasswordInput(userID, &oldPassword, &newPassword)
	if err != nil {
		return
	}

	user, err := s.FindUserByID(ctx, userID)
	if user == nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return app.InputError{"old_password": ErrInvalidPassword.Error()}
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), app.DefaultBCryptCost)
	if err != nil {
		return err
	}

	return s.db.UpdateUserPassword(ctx, user.ID, string(newPasswordHash))
}

// ValidateChangeUserPasswordInput ...
func ValidateChangeUserPasswordInput(userID int64, oldP, newP *string) (err error) {
	in := &ChangeUserPasswordInput{
		UserID:      userID,
		OldPassword: *oldP,
		NewPassword: *newP,
	}

	in.OldPassword = strings.TrimSpace(in.OldPassword)
	in.NewPassword = strings.TrimSpace(in.NewPassword)

	err = validateInput(changePasswordInputRules, in)
	if err != nil {
		return
	}

	*oldP, *newP = in.OldPassword, in.NewPassword
	return nil
}

// ChangeUserPasswordInput defines the input for changing a user's password.
type ChangeUserPasswordInput struct {
	UserID      int64
	OldPassword string
	NewPassword string
}
