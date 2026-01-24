package service

import (
	"context"

	"github.com/gopl-dev/server/app/ds"
)

// FilterBooks ...
func (s *Service) FilterBooks(ctx context.Context, f ds.BooksFilter) (data []ds.Book, count int, err error) {
	ctx, span := s.tracer.Start(ctx, "FilterBooks")
	defer span.End()

	return s.db.FilterBooks(ctx, f)
}
