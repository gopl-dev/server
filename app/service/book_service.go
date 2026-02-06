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
	"github.com/gopl-dev/server/email"
)

var (
	// ErrCoverIsNotABookCover indicates that the provided file
	// cannot be used as a book cover because it is not marked
	// or classified as a book cover.
	ErrCoverIsNotABookCover = errors.New("cover: not a book cover")

	// ErrCoverBelongsToAnotherUser indicates that the provided
	// cover file is owned by a different user and therefore
	// cannot be attached to the current user's book.
	ErrCoverBelongsToAnotherUser = errors.New("cover: not owner")

	// ErrBookIsNotUnderReview is returned when an operation requires a book
	// to be in the "under review" state.
	ErrBookIsNotUnderReview = errors.New("book is not under review")
)

// FilterBooks retrieves a paginated list of books matching the given filter.
func (s *Service) FilterBooks(ctx context.Context, f ds.BooksFilter) (data []ds.Book, count int, err error) {
	ctx, span := s.tracer.Start(ctx, "FilterBooks")
	defer span.End()

	return s.db.FilterBooks(ctx, f)
}

// CreateBook handles the transactional creation of a book, with its base entity and logs.
func (s *Service) CreateBook(ctx context.Context, book *ds.Book) (err error) {
	ctx, span := s.tracer.Start(ctx, "CreateBook")
	defer span.End()

	book.Summary, err = app.MarkdownToHTML(book.SummaryRaw)
	if err != nil {
		return
	}
	book.Description, err = app.MarkdownToHTML(book.DescriptionRaw)
	if err != nil {
		return
	}

	err = ValidateCreate(book)
	if err != nil {
		return err
	}

	err = s.resolveBookCover(ctx, book, false)
	if err != nil {
		return err
	}

	book.Topics, err = s.normalizeTopics(ctx, book.Topics, ds.EntityTypeBook, 1)
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

		err = s.AttachTopics(ctx, book.ID, book.Topics)
		if err != nil {
			return
		}

		if !book.CoverFileID.IsNil() {
			err = s.db.CommitFile(ctx, book.CoverFileID)
			if err != nil {
				return
			}
		}

		return nil
	})
}

// ApproveNewBook approves a newly submitted book.
func (s *Service) ApproveNewBook(ctx context.Context, book *ds.Book) (err error) {
	ctx, span := s.tracer.Start(ctx, "ApproveNewBook")
	defer span.End()

	if book.Status.Not(ds.EntityStatusUnderReview) {
		err = ErrBookIsNotUnderReview
		return
	}

	user := ds.UserFromContext(ctx)
	if user == nil {
		err = app.ErrUnauthorized()
		return
	}

	if !user.IsAdmin {
		err = app.ErrUnauthorized()
		return
	}

	err = s.db.WithTx(ctx, func(ctx context.Context) (err error) {
		err = s.ChangeEntityStatus(ctx, book.ID, ds.EntityStatusApproved)
		if err != nil {
			return
		}

		return s.LogBookApproved(ctx, user.ID, book)
	})
	if err != nil {
		return
	}

	owner, err := s.FindUserByID(ctx, book.OwnerID)
	if err != nil {
		return
	}

	return email.Send(owner.Email, email.BookApproved{
		BookName: book.Title,
		Username: owner.Username,
		PublicID: book.PublicID,
	})
}

// RejectNewBook rejects a newly submitted book.
func (s *Service) RejectNewBook(ctx context.Context, note string, book *ds.Book) (err error) {
	ctx, span := s.tracer.Start(ctx, "RejectNewBook")
	defer span.End()

	if book.Status.Not(ds.EntityStatusUnderReview) {
		err = ErrBookIsNotUnderReview
		return
	}

	user := ds.UserFromContext(ctx)
	if user == nil {
		err = app.ErrUnauthorized()
		return
	}

	if !user.IsAdmin {
		err = app.ErrUnauthorized()
		return
	}

	err = s.db.WithTx(ctx, func(ctx context.Context) (err error) {
		err = s.ChangeEntityStatus(ctx, book.ID, ds.EntityStatusRejected)
		if err != nil {
			return
		}

		return s.LogBookRejected(ctx, user.ID, note, book)
	})
	if err != nil {
		return
	}

	owner, err := s.FindUserByID(ctx, book.OwnerID)
	if err != nil {
		return
	}

	return email.Send(owner.Email, email.BookRejected{
		Note:     note,
		BookName: book.Title,
		Username: owner.Username,
	})
}

// resolveBookCover validates and normalizes the book cover file reference.
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

	newBook.Summary, err = app.MarkdownToHTML(newBook.SummaryRaw)
	if err != nil {
		return
	}
	newBook.Description, err = app.MarkdownToHTML(newBook.DescriptionRaw)
	if err != nil {
		return
	}

	err = ValidateCreate(newBook)
	if err != nil {
		return
	}

	newBook.Topics, err = s.normalizeTopics(ctx, newBook.Topics, ds.EntityTypeBook, 1)
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
			// TODO create change request, so we can have a diff for history

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

			if isRenameOnly(diff) {
				return s.LogEntityRenamed(ctx, book.Title, newBook.Entity)
			}

			return s.LogEntityUpdated(ctx, newBook.Entity)
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

// isRenameOnly reports whether the update data represents a title-only change.
func isRenameOnly(data map[string]any) bool {
	_, ok := data["title"]
	return ok && len(data) == 1
}
