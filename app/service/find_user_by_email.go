package service

import (
	"context"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
)

var findUserByEmailInputRules = z.Shape{
	"email": emailInputRules,
}

// FindUserByEmail retrieves a user record from the database by their email address.
func (s *Service) FindUserByEmail(ctx context.Context, email string) (user *ds.User, err error) {
	ctx, span := s.tracer.Start(ctx, "FindUserByEmail")
	defer span.End()

	err = ValidateFindUserByEmailInput(&email)
	if err != nil {
		return
	}

	return s.db.FindUserByEmail(ctx, email)
}

// ValidateFindUserByEmailInput ...
func ValidateFindUserByEmailInput(email *string) (err error) {
	in := &FindUserByEmailInput{
		Email: *email,
	}

	in.Email = strings.TrimSpace(in.Email)

	err = validateInput(findUserByEmailInputRules, in)
	if err != nil {
		return
	}

	*email = in.Email
	return nil
}

// FindUserByEmailInput ...
type FindUserByEmailInput struct {
	Email string
}
