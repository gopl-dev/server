package service

import (
	"context"
	"errors"
	"strings"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/test/factory/random"
)

// CreateEntity validates and persists a new entity in the database.
// It automatically handles URLName generation: if empty, it creates a slug from the Title.
// If the URLName already exists, it appends a random suffix to ensure uniqueness.
func (s *Service) CreateEntity(ctx context.Context, e *ds.Entity) error {
	ctx, span := s.tracer.Start(ctx, "CreateEntity")
	defer span.End()

	user := ds.UserFromContext(ctx)
	if user == nil {
		return app.ErrUnauthorized()
	}

	// resolve URLName
	if strings.TrimSpace(e.URLName) == "" {
		e.URLName = app.Slug(e.Title)
		if strings.TrimSpace(e.URLName) == "" {
			e.URLName = random.String(10) //nolint:mnd
		}
	}

resolveURLName:
	existing, err := s.db.FindEntityByURLName(ctx, e.URLName)
	if existing != nil {
		e.URLName += "-" + random.String(5) //nolint:mnd
		goto resolveURLName
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

	return s.db.CreateEntity(ctx, e)
}
