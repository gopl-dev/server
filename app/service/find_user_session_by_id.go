package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/gopl-dev/server/app/ds"
)

// FindUserSessionByID retrieves a user session from the database using its ID.
func (s *Service) FindUserSessionByID(ctx context.Context, id uuid.UUID) (sess *ds.UserSession, err error) {
	ctx, span := s.tracer.Start(ctx, "FindUserSessionByID")
	defer span.End()

	in := &FindUserSessionByIDInput{ID: id}
	err = Normalize(in)
	if err != nil {
		return
	}

	return s.db.FindUserSessionByID(ctx, id)
}

// FindUserSessionByIDInput ...
type FindUserSessionByIDInput struct {
	ID uuid.UUID
}

// Sanitize ...
func (in *FindUserSessionByIDInput) Sanitize() {}

// Validate ...
func (in *FindUserSessionByIDInput) Validate() error {
	return nil
}
