package service

import (
	"context"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
)

var findUserByEmailInputRules = z.Shape{
	"Email": emailInputRules,
}

// FindUserByEmail retrieves a user record from the database by their email address.
func (s *Service) FindUserByEmail(ctx context.Context, email string) (user *ds.User, err error) {
	ctx, span := s.tracer.Start(ctx, "FindUserByEmail")
	defer span.End()

	in := &FindUserByEmailInput{Email: email}
	err = Normalize(in)
	if err != nil {
		return
	}

	return s.db.FindUserByEmail(ctx, in.Email)
}

// FindUserByEmailInput ...
type FindUserByEmailInput struct {
	Email string
}

// Sanitize ...
func (in *FindUserByEmailInput) Sanitize() {
	in.Email = strings.TrimSpace(in.Email)
}

// Validate ...
func (in *FindUserByEmailInput) Validate() error {
	return validateInput(findUserByEmailInputRules, in)
}
