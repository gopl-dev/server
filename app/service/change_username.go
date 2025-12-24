package service

import (
	"context"
	"errors"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/repo"
	"golang.org/x/crypto/bcrypt"
)

var changeUsernameInputRules = z.Shape{
	"UserID":      userIDInputRules,
	"NewUsername": usernameInputRules,
	"Password":    z.String().Required(),
}

// ChangeUsername handles the logic for changing a user's username.
func (s *Service) ChangeUsername(ctx context.Context, in ChangeUsernameInput) (err error) {
	ctx, span := s.tracer.Start(ctx, "ChangeUsername")
	defer span.End()

	err = Normalize(&in)
	if err != nil {
		return
	}

	user, err := s.db.FindUserByID(ctx, in.UserID)
	if err != nil {
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password))
	if err != nil {
		return app.InputError{"password": "Incorrect password"}
	}

	existingUser, err := s.db.FindUserByUsername(ctx, in.NewUsername)
	if errors.Is(err, repo.ErrUserNotFound) {
		err = nil
	}
	if err != nil {
		return
	}
	if existingUser != nil {
		return app.InputError{"username": UsernameAlreadyTaken}
	}

	return s.db.UpdateUsername(ctx, user.ID, in.NewUsername)
}

// ChangeUsernameInput ...
type ChangeUsernameInput struct {
	UserID      int64
	NewUsername string
	Password    string
}

// Sanitize ...
func (in *ChangeUsernameInput) Sanitize() {
	in.NewUsername = strings.TrimSpace(in.NewUsername)
}

// Validate ...
func (in *ChangeUsernameInput) Validate() error {
	return validateInput(changeUsernameInputRules, in)
}
