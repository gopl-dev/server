package service

import (
	"context"

	"github.com/gopl-dev/server/app/ds"
)

// CreateBook handles the transactional creation of a book, with its base entity and logs.
func (s *Service) CreateBook(ctx context.Context, book *ds.Book) error {
	ctx, span := s.tracer.Start(ctx, "CreateBook")
	defer span.End()

	err := ValidateCreate(book)
	if err != nil {
		return err
	}

	return s.db.WithTx(ctx, func(ctx context.Context) (err error) {
		err = s.CreateEntity(ctx, &book.Entity)
		if err != nil {
			return
		}

		err = s.db.CreateBook(ctx, book)
		if err != nil {
			return
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
