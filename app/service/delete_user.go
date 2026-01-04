package service

import (
	"context"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"golang.org/x/crypto/bcrypt"
)

var deleteUserInputRules = z.Shape{
	"UserID":   ds.IDInputRules,
	"Password": z.String().Required(),
}

// DeleteUser handles the logic for soft-deleting a user account.
func (s *Service) DeleteUser(ctx context.Context, userID ds.ID, password string) (err error) {
	ctx, span := s.tracer.Start(ctx, "DeleteUser")
	defer span.End()

	in := DeleteUserInput{
		UserID:   userID,
		Password: password,
	}
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

	err = s.db.DeleteUser(ctx, user.ID)
	if err != nil {
		return
	}

	err = s.db.DeleteSessionsByUserID(ctx, user.ID)
	if err != nil {
		return
	}

	return
}

// DeleteUserInput ...
type DeleteUserInput struct {
	UserID   ds.ID
	Password string
}

// Sanitize ...
func (in *DeleteUserInput) Sanitize() {
}

// Validate ...
func (in *DeleteUserInput) Validate() error {
	return validateInput(deleteUserInputRules, in)
}
