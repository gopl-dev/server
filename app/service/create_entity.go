package service

import (
	"context"
	"errors"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/test/factory/random"
)

// CreateEntity validates and creates a new entity in the database.
func (s *Service) CreateEntity(ctx context.Context, e *ds.Entity) error {
	ctx, span := s.tracer.Start(ctx, "CreateEntity")
	defer span.End()

	user := ds.UserFromContext(ctx)
	if user == nil {
		return app.ErrUnauthorized()
	}

	e.SetPublicID()

resolvePublicID:
	existing, err := s.db.FindEntityByPublicID(ctx, e.PublicID, e.Type)
	if existing != nil {
		e.PublicID += "-" + random.String(5) //nolint:mnd
		goto resolvePublicID
	}
	if errors.Is(err, repo.ErrEntityNotFound) {
		err = nil
	}
	if err != nil {
		return err
	}

	// resolve visibility
	e.Status = ds.EntityStatusUnderReview
	// for admins and private entities set status to approved
	if user.IsAdmin || e.Visibility.Is(ds.EntityVisibilityPrivate) {
		e.Status = ds.EntityStatusApproved
	}

	err = ValidateCreate(e)
	if err != nil {
		return err
	}

	err = s.db.CreateEntity(ctx, e)
	if err != nil {
		return err
	}

	return s.LogEntityCreated(ctx, e)
}
