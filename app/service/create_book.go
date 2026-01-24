package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
)

var (
	// ErrCoverIsNotABookCover ...
	ErrCoverIsNotABookCover = errors.New("cover: not a book cover")

	// ErrCoverBelongsToAnotherUser ...
	ErrCoverBelongsToAnotherUser = errors.New("cover: not owner")
)

// CreateBook handles the transactional creation of a book, with its base entity and logs.
func (s *Service) CreateBook(ctx context.Context, book *ds.Book) error {
	ctx, span := s.tracer.Start(ctx, "CreateBook")
	defer span.End()

	err := ValidateCreate(book)
	if err != nil {
		return err
	}

	if !book.CoverFileID.IsNil() {
		cover, err := s.GetFileByID(ctx, book.CoverFileID)
		if errors.Is(err, repo.ErrFileNotFound) {
			book.CoverFileID = ds.NilID
			book.PreviewFileID = ds.NilID
			err = nil
		}
		if err != nil {
			return fmt.Errorf("get cover: %w", err)
		}

		if !cover.IsOwner(book.OwnerID) {
			return ErrCoverBelongsToAnotherUser
		}

		if !cover.IsBookCover() {
			return ErrCoverIsNotABookCover
		}

		book.PreviewFileID = book.CoverFileID
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
