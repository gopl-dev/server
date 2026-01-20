package service

import (
	"context"

	"github.com/gopl-dev/server/app/ds"
)

// FilterUsers ...
func (s *Service) FilterUsers(ctx context.Context, f ds.UsersFilter) (data []ds.User, count int, err error) {
	ctx, span := s.tracer.Start(ctx, "FilterUsers")
	defer span.End()

	return s.db.FilterUsers(ctx, f)
}
