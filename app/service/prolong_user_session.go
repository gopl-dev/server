package service

import (
	"context"

	"github.com/google/uuid"
)

// ProlongUserSession updates the expiration time of an existing user session in the database.
func (s *Service) ProlongUserSession(ctx context.Context, id uuid.UUID) (err error) {
	ctx, span := s.tracer.Start(ctx, "ProlongUserSession")
	defer span.End()

	in := &ProlongUserSessionInput{ID: id}
	err = Normalize(in)
	if err != nil {
		return
	}

	return s.db.ProlongUserSession(ctx, id)
}

// ProlongUserSessionInput ...
type ProlongUserSessionInput struct {
	ID uuid.UUID
}

// Sanitize ...
func (in *ProlongUserSessionInput) Sanitize() {}

// Validate ...
func (in *ProlongUserSessionInput) Validate() error {
	return nil
}
