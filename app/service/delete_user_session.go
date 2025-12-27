package service

import (
	"context"

	"github.com/gopl-dev/server/app/ds"
)

// DeleteUserSession removes a user session record from the database using its ID.
func (s *Service) DeleteUserSession(ctx context.Context, id ds.ID) (err error) {
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
	ID ds.ID
}

// Sanitize ...
func (in *DeleteUserSessionInput) Sanitize() {}

// Validate ...
func (in *DeleteUserSessionInput) Validate() error {
	return nil
}
