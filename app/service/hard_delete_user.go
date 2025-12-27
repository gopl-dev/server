package service

import (
	"context"

	z "github.com/Oudwins/zog"
)

var hardDeleteUserInputRules = z.Shape{
	"UserID": userIDInputRules,
}

// HardDeleteUser handles the logic for deleting a user account and relations.
func (s *Service) HardDeleteUser(ctx context.Context, userID int64) (err error) {
	ctx, span := s.tracer.Start(ctx, "DeleteUser")
	defer span.End()

	// sessions
	err = s.db.DeleteSessionsByUserID(ctx, userID)
	if err != nil {
		return
	}

	// email confirmations
	err = s.db.DeleteEmailConfirmationByUser(ctx, userID)
	if err != nil {
		return
	}

	// password resets
	err = s.db.DeletePasswordResetTokensByUser(ctx, userID)
	if err != nil {
		return
	}

	// email change requests
	err = s.db.DeleteChangeEmailRequestsByUser(ctx, userID)
	if err != nil {
		return
	}

	// user
	return s.db.HardDeleteUser(ctx, userID)
}

// HardDeleteUserInput ...
type HardDeleteUserInput struct {
	UserID int64
}

// Sanitize ...
func (in *HardDeleteUserInput) Sanitize() {
}

// Validate ...
func (in *HardDeleteUserInput) Validate() error {
	return validateInput(hardDeleteUserInputRules, in)
}
