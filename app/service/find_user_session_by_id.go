package service

import (
	"context"

	"github.com/gopl-dev/server/app/ds"
)

// FindUserSessionByID retrieves a user session from the database using its ID.
func (s *Service) FindUserSessionByID(ctx context.Context, id ds.ID) (sess *ds.UserSession, err error) {
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
	ID ds.ID
}

// Sanitize ...
func (in *FindUserSessionByIDInput) Sanitize() {}

// Validate ...
func (in *FindUserSessionByIDInput) Validate() error {
	return nil
}
