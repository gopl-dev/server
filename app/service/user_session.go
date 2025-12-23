package service

import (
	"context"

	"github.com/google/uuid"
)

// ProlongUserSession updates the expiration time of an existing user session in the database.
func (s *Service) ProlongUserSession(ctx context.Context, id uuid.UUID) (err error) {
	ctx, span := s.tracer.Start(ctx, "ProlongUserSession")
	defer span.End()

	err = s.db.ProlongUserSession(ctx, id)
	return
}

// DeleteUserSession removes a user session record from the database using its ID.
func (s *Service) DeleteUserSession(ctx context.Context, id uuid.UUID) (err error) {
	ctx, span := s.tracer.Start(ctx, "DeleteUserSession")
	defer span.End()

	err = s.db.DeleteUserSession(ctx, id)
	return
}
