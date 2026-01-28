package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
)

var (
	// ErrCoverIsNotABookCover ...
	ErrCoverIsNotABookCover = errors.New("cover: not a book cover")

	// ErrCoverBelongsToAnotherUser ...
	ErrCoverBelongsToAnotherUser = errors.New("cover: not owner")
)

// FilterBooks ...
func (s *Service) FilterBooks(ctx context.Context, f ds.BooksFilter) (data []ds.Book, count int, err error) {
	ctx, span := s.tracer.Start(ctx, "FilterBooks")
	defer span.End()

	return s.db.FilterBooks(ctx, f)
}

// CreateBook handles the transactional creation of a book, with its base entity and logs.
func (s *Service) CreateBook(ctx context.Context, book *ds.Book) error {
	ctx, span := s.tracer.Start(ctx, "CreateBook")
	defer span.End()

	err := ValidateCreate(book)
	if err != nil {
		return err
	}

	err = s.resolveBookCover(ctx, book, false)
	if err != nil {
		return err
	}

	return s.db.WithTx(ctx, func(ctx context.Context) (err error) {
		err = s.CreateEntity(ctx, book.Entity)
		if err != nil {
			return
		}

		err = s.db.CreateBook(ctx, book)
		if err != nil {
			return
		}

		if !book.CoverFileID.IsNil() {
			err = s.db.CommitFile(ctx, book.CoverFileID)
			if err != nil {
				return
			}
		}

		log := &ds.EntityChangeLog{
			ID:        ds.NewID(),
			EntityID:  book.ID,
			UserID:    book.OwnerID,
			Action:    ds.ActionCreate,
			CreatedAt: book.CreatedAt,
		}

		return s.db.CreateEntityChangeLog(ctx, log)
	})
}

func (s *Service) resolveBookCover(ctx context.Context, book *ds.Book, edit bool) (err error) {
	if book.CoverFileID.IsNil() {
		return
	}

	cover, err := s.GetFileByID(ctx, book.CoverFileID)
	if errors.Is(err, repo.ErrFileNotFound) {
		book.CoverFileID = ds.NilID
		book.PreviewFileID = ds.NilID
		return nil
	}
	if err != nil {
		return fmt.Errorf("get cover: %w", err)
	}

	if !cover.IsOwner(book.OwnerID) && !edit {
		return ErrCoverBelongsToAnotherUser
	}

	if !cover.IsBookCover() {
		return ErrCoverIsNotABookCover
	}

	book.PreviewFileID = book.CoverFileID
	return nil
}

// UpdateBook updates an existing book by its identifier.
//
// For admin users, changes are applied immediately.
//
// For non-admin users, a pending entity change request is created instead,
// and the update must be reviewed before being applied.
//
// The method returns the resulting revision number. For direct admin
// updates, the revision is always 0.
func (s *Service) UpdateBook(ctx context.Context, id ds.ID, newBook *ds.Book) (revision int, err error) {
	ctx, span := s.tracer.Start(ctx, "UpdateBook")
	defer span.End()

	user := ds.UserFromContext(ctx)
	if user == nil {
		err = app.ErrUnauthorized()
		return
	}

	err = ValidateCreate(newBook)
	if err != nil {
		return
	}

	book, err := s.GetBookByID(ctx, id)
	if err != nil {
		return
	}

	newBook.ID = book.ID
	newBook.OwnerID = book.OwnerID

	diff, ok := makeDiff(book, newBook)
	if !ok {
		return
	}

	// If an authorized user makes changes, apply them right away and mark them as well-done.
	// Note: "user.IsAdmin" is not how we're going to handle this long term — proper RBAC will be implemented.
	// But for now, we need some kind of raw authority check.
	if user.IsAdmin {
		err = s.db.WithTx(ctx, func(ctx context.Context) (err error) {
			if newBook.CoverFileID != book.CoverFileID {
				err = s.resolveBookCover(ctx, newBook, true)
				if err != nil {
					return
				}

				if !book.CoverFileID.IsNil() {
					err = s.DeleteFile(ctx, book.CoverFileID)
					if errors.Is(err, repo.ErrFileNotFound) {
						err = nil
					}
					if err != nil {
						return err
					}
				}

				err = s.db.CommitFile(ctx, newBook.CoverFileID)
				if err != nil {
					return
				}
			}

			err = s.db.UpdateEntity(ctx, newBook.Entity)
			if err != nil {
				return err
			}

			err = s.db.UpdateBook(ctx, newBook)
			if err != nil {
				return err
			}

			log := &ds.EntityChangeLog{
				ID:        ds.NewID(),
				EntityID:  book.ID,
				UserID:    user.ID,
				Diff:      diff,
				Action:    ds.ActionEdit,
				CreatedAt: book.CreatedAt,
			}

			return s.db.CreateEntityChangeLog(ctx, log)
		})

		return 0, err
	}

	req := &ds.EntityChangeRequest{
		ID:         ds.NewID(),
		EntityID:   book.ID,
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

// makeDiff compares two DataProvider states and returns a diff map that contains
// only fields whose values changed in newData compared to oldData.
//
// The comparison is performed only for keys present in oldData (acts as the allowed
// field set). The diff values contain the new values from newData. If no changes
// are found, diff is nil and hasDiff is false.
// TODO find a home for this awesome function.
func makeDiff(oldData, newData ds.DataProvider) (diff map[string]any, hasDiff bool) {
	oldMap := oldData.Data()
	newMap := newData.Data()

	for key, oldVal := range oldMap {
		newVal, ok := newMap[key]
		if !ok {
			continue // no key - no changes
		}

		if oldVal == nil || newVal == nil {
			if oldVal == newVal {
				continue
			}
		} else if reflect.DeepEqual(oldVal, newVal) {
			continue
		}

		if diff == nil {
			diff = make(map[string]any)
		}
		diff[key] = newVal
	}

	return diff, diff != nil
}

// hasDiff reports whether newData differs from oldData.
func hasDiff(oldData, newData map[string]any) bool {
	if len(oldData) != len(newData) {
		return true
	}

	for key, oldVal := range oldData {
		newVal, ok := newData[key]
		if !ok {
			return true
		}

		if oldVal == nil || newVal == nil {
			if oldVal == newVal {
				continue
			}
		} else if reflect.DeepEqual(oldVal, newVal) {
			continue
		}

		return true
	}

	return false
}
