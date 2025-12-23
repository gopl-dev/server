package service

import (
	"context"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	"github.com/gopl-dev/server/app/ds"
)

var findUserSessionByIDInputRules = z.Shape{
	"ID": z.CustomFunc(func(id *string, _ z.Ctx) bool {
		return uuid.Validate(*id) == nil
	}, z.Message("Invalid UUID")),
}

// FindUserSessionByID retrieves a user session from the database using its ID.
func (s *Service) FindUserSessionByID(ctx context.Context, id uuid.UUID) (sess *ds.UserSession, err error) {
	ctx, span := s.tracer.Start(ctx, "FindUserSessionByID")
	defer span.End()

	err = ValidateFindUserSessionByIDInput(id.String())
	if err != nil {
		return
	}

	return s.db.FindUserSessionByID(ctx, id)
}

// ValidateFindUserSessionByIDInput ...
func ValidateFindUserSessionByIDInput(id string) (err error) {
	in := &FindUserSessionByIDInput{
		ID: id,
	}

	return validateInput(findUserSessionByIDInputRules, in)
}

// FindUserSessionByIDInput ...
type FindUserSessionByIDInput struct {
	ID string
}
