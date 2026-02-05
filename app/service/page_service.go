package service

import (
	"context"
	"errors"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
)

// GetPageByPublicID retrieves a page by its public identifier.
func (s *Service) GetPageByPublicID(ctx context.Context, id string) (*ds.Page, error) {
	ctx, span := s.tracer.Start(ctx, "GetPageByPublicID")
	defer span.End()

	return s.db.GetPageByPublicID(ctx, id)
}

// CreatePage handles creation of a page.
func (s *Service) CreatePage(ctx context.Context, page *ds.Page) (err error) {
	ctx, span := s.tracer.Start(ctx, "CreateBook")
	defer span.End()

	page.Content, err = app.MarkdownToHTML(page.ContentRaw)
	if err != nil {
		return
	}

	err = ValidateCreate(page)
	if err != nil {
		return err
	}

	existing, err := s.GetPageByPublicID(ctx, page.PublicID)
	if errors.Is(err, repo.ErrPageNotFound) {
		err = nil
	}
	if err != nil {
		return err
	}
	if existing != nil {
		return app.NewInputError(
			"public_id",
			"Page with this Public ID '%s' already exists.",
			page.PublicID,
		)
	}

	return s.db.WithTx(ctx, func(ctx context.Context) (err error) {
		err = s.CreateEntity(ctx, page.Entity)
		if err != nil {
			return
		}

		err = s.db.CreatePage(ctx, page)
		if err != nil {
			return
		}
		return nil
	})
}

// UpdatePage updates an existing page identified by its public ID.
//
// If the caller is an admin, changes are applied immediately and recorded
// in the entity change log. For non-admin users, a pending change request
// is created instead, and the update must be reviewed before being applied.
//
// The method returns the resulting revision number. For direct admin
// updates, the revision is always 0.
func (s *Service) UpdatePage(ctx context.Context, id string, newPage *ds.Page) (revision int, err error) {
	ctx, span := s.tracer.Start(ctx, "UpdatePage")
	defer span.End()

	user := ds.UserFromContext(ctx)
	if user == nil {
		err = app.ErrUnauthorized()
		return
	}

	newPage.Content, err = app.MarkdownToHTML(newPage.ContentRaw)
	if err != nil {
		return
	}

	err = ValidateCreate(newPage)
	if err != nil {
		return
	}

	page, err := s.GetPageByPublicID(ctx, id)
	if err != nil {
		return
	}

	newPage.ID = page.ID
	newPage.OwnerID = page.OwnerID

	diff, ok := makeDiff(page, newPage)
	if !ok {
		return
	}

	if user.IsAdmin {
		err = s.db.WithTx(ctx, func(ctx context.Context) (err error) {
			// TODO create change request, so we can have a diff
			err = s.db.UpdateEntity(ctx, newPage.Entity)
			if err != nil {
				return err
			}

			err = s.db.UpdatePage(ctx, newPage)
			if err != nil {
				return err
			}

			if isRenameOnly(diff) {
				return s.LogEntityRenamed(ctx, page.Title, newPage.Entity)
			}

			return s.LogEntityUpdated(ctx, newPage.Entity)
		})

		return 0, err
	}

	req := &ds.EntityChangeRequest{
		ID:         ds.NewID(),
		EntityID:   page.ID,
		UserID:     user.ID,
		Status:     ds.EntityChangePending,
		Diff:       diff,
		Message:    "",
		Revision:   0,
		ReviewerID: nil,
		ReviewedAt: nil,
		ReviewNote: "",
		CreatedAt:  time.Now(),
		UpdatedAt:  nil,
	}

	err = s.UpdateEntityChangeRequest(ctx, req)
	if err != nil {
		return
	}

	return req.Revision, nil
}
