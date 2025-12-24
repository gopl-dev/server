package service

import (
	"context"

	"github.com/google/uuid"
)

// DeleteUserSession removes a user session record from the database using its ID.
func (s *Service) DeleteUserSession(ctx context.Context, id uuid.UUID) (err error) {
	ctx, span := s.tracer.Start(ctx, "DeleteUserSession")
	defer span.End()

	in := &DeleteUserSessionInput{ID: id}
	err = Normalize(in)
	if err != nil {
		return
	}

	return s.db.DeleteUserSession(ctx, in.ID)
}

// DeleteUserSessionInput ...
type DeleteUserSessionInput struct {
	ID uuid.UUID
}

// Sanitize ...
func (in *DeleteUserSessionInput) Sanitize() {}

// Validate ...
func (in *DeleteUserSessionInput) Validate() error {
	return nil
}
