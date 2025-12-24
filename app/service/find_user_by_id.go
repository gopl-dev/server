package service

import (
	"context"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
)

var findUserByIDInputRules = z.Shape{
	"ID": z.Int64().Required(z.Message("ID is required")),
}

// FindUserByID retrieves a user record from the database by their ID.
func (s *Service) FindUserByID(ctx context.Context, id int64) (user *ds.User, err error) {
	ctx, span := s.tracer.Start(ctx, "FindUserByID")
	defer span.End()

	in := &FindUserByIDInput{ID: id}
	err = Normalize(in)
	if err != nil {
		return
	}

	return s.db.FindUserByID(ctx, in.ID)
}

// FindUserByIDInput ...
type FindUserByIDInput struct {
	ID int64
}

// Sanitize ...
func (in *FindUserByIDInput) Sanitize() {}

// Validate ...
func (in *FindUserByIDInput) Validate() error {
	return validateInput(findUserByIDInputRules, in)
}
