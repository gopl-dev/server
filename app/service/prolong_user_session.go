package service

import (
	"context"

	"github.com/gopl-dev/server/app/ds"
)

// ProlongUserSession updates the expiration time of an existing user session in the database.
func (s *Service) ProlongUserSession(ctx context.Context, id ds.ID) (err error) {
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
	ID ds.ID
}

// Sanitize ...
func (in *ProlongUserSessionInput) Sanitize() {}

// Validate ...
func (in *ProlongUserSessionInput) Validate() error {
	return nil
}
