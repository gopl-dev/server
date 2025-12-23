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

	err = ValidateFindUserByIDInput(id)
	if err != nil {
		return
	}

	return s.db.FindUserByID(ctx, id)
}

// ValidateFindUserByIDInput ...
func ValidateFindUserByIDInput(id int64) (err error) {
	in := &FindUserByIDInput{
		ID: id,
	}

	return validateInput(findUserByIDInputRules, in)
}

// FindUserByIDInput defines the input for changing a user's password.
type FindUserByIDInput struct {
	ID int64
}
