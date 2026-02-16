package service

import (
	"context"
	"errors"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/email"
)

// GetPageByPublicID retrieves a page by its public identifier.
func (s *Service) GetPageByPublicID(ctx context.Context, id string) (*ds.Page, error) {
	ctx, span := s.tracer.Start(ctx, "GetPageByPublicID")
	defer span.End()

	return s.db.GetPageByPublicID(ctx, id)
}

// GetPageByID retrieves a page by its ID from the database.
func (s *Service) GetPageByID(ctx context.Context, id ds.ID) (*ds.Page, error) {
	ctx, span := s.tracer.Start(ctx, "GetBookByID")
	defer span.End()

	return s.db.GetPageByID(ctx, id)
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
// in the entity change log.
func (s *Service) UpdatePage(ctx context.Context, id string, newPage *ds.Page) (req *ds.EntityChangeRequest, err error) {
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

	req = &ds.EntityChangeRequest{
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

	if user.IsAdmin {
		changes, err := makeChangesDiff(page, diff)
		if err != nil {
			return nil, err
		}
		err = s.ApplyChangesToPage(ctx, changes, req, false)
		if err != nil {
			return nil, err
		}
	}

	return req, nil
}

// ApplyChangesToPage applies approved changes from a change request to a page entity.
func (s *Service) ApplyChangesToPage(ctx context.Context, changes []ChangeDiff, req *ds.EntityChangeRequest, sendNotification bool) (err error) {
	ctx, span := s.tracer.Start(ctx, "ApplyChangesToPage")
	defer span.End()

	user := ds.UserFromContext(ctx)
	if user == nil {
		return app.ErrUnauthorized()
	}

	page, err := s.GetPageByID(ctx, req.EntityID)
	if err != nil {
		return
	}

	author, err := s.GetUserByID(ctx, req.UserID)
	if err != nil {
		return
	}

	entityData, data, err := normalizeDataFromChangeRequest(page, req.Diff)
	if err != nil {
		return
	}

	err = s.db.WithTx(ctx, func(ctx context.Context) (err error) {
		err = s.ApplyChangesToEntity(ctx, page.Entity, entityData)
		if err != nil {
			return
		}

		if len(data) > 0 {
			err = s.db.ApplyChangesToPage(ctx, req.EntityID, data)
			if err != nil {
				return
			}
		}

		err = s.CommitChangeRequest(ctx, req)
		if err != nil {
			return err
		}

		if isRenameOnly(req.Diff) {
			return s.LogEntityRenamed(ctx, req.UserID, req.EntityID, page.Title, entityData["title"])
		}

		err = s.LogEntityUpdated(ctx, req.UserID, req.EntityID, page.Title, changes)
		if err != nil {
			return
		}

		return nil
	})
	if err != nil {
		return
	}

	if publicID, ok := entityData["public_id"]; ok {
		page.PublicID = app.String(publicID)
	}

	if sendNotification {
		err = email.Send(author.Email, email.ChangesApproved{
			Username:    author.Username,
			EntityTitle: page.Title,
			ViewURL:     page.PublicID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
