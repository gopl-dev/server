package service

import (
	"context"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
)

var filterUsersInputRules = z.Shape{
	"UserID": ds.IDInputRules,
}

// FilterUsers ...
func (s *Service) FilterUsers(ctx context.Context, f ds.UsersFilter) (data []ds.User, count int, err error) {
	ctx, span := s.tracer.Start(ctx, "FilterUsers")
	defer span.End()

	return s.db.FilterUsers(ctx, f)
}

// FilterUsersInput ...
type FilterUsersInput struct {
	UserID int64
}

// Sanitize ...
func (in *FilterUsersInput) Sanitize() {
}

// Validate ...
func (in *FilterUsersInput) Validate() error {
	return validateInput(filterUsersInputRules, in)
}
