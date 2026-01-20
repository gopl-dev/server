package service

import (
	"context"
	"errors"

	"github.com/gopl-dev/server/app/ds"
)

// ErrInvalidRefID ...
var ErrInvalidRefID = errors.New("id must be UUID or string")

// GetBookByID ...
func (s *Service) GetBookByID(ctx context.Context, id ds.ID) (*ds.Book, error) {
	ctx, span := s.tracer.Start(ctx, "GetBookByID")
	defer span.End()

	return s.db.GetBookByID(ctx, id)
}

// GetBookByPublicID ...
func (s *Service) GetBookByPublicID(ctx context.Context, publicID string) (*ds.Book, error) {
	ctx, span := s.tracer.Start(ctx, "GetBookByPublicID")
	defer span.End()

	return s.db.GetBookByPublicID(ctx, publicID)
}

// GetBookByRef returns a book by a reference of unknown type.
//
// The reference may be either:
//   - ds.ID (internal UUID-based identifier), or
//   - string, representing either a UUID or a public identifier (e.g. "book_ABCXYZ").
func (s *Service) GetBookByRef(ctx context.Context, ref any) (*ds.Book, error) {
	ctx, span := s.tracer.Start(ctx, "GetBookByRef")
	defer span.End()

	id, ok := ref.(ds.ID)
	if ok {
		return s.db.GetBookByID(ctx, id)
	}

	idStr, ok := ref.(string)
	if ok {
		id, err := ds.ParseID(idStr)
		if err == nil {
			return s.db.GetBookByID(ctx, id)
		}

		return s.db.GetBookByPublicID(ctx, idStr)
	}

	return nil, ErrInvalidRefID
}
