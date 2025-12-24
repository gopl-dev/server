package service

import (
	"context"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

var createUserSessionInputRules = z.Shape{
	"UserID": userIDInputRules,
}

// CreateUserSession creates a new user session object.
func (s *Service) CreateUserSession(ctx context.Context, userID int64) (sess *ds.UserSession, err error) {
	ctx, span := s.tracer.Start(ctx, "CreateUserSession")
	defer span.End()

	in := &CreateUserSessionInput{UserID: userID}
	err = Normalize(in)
	if err != nil {
		return
	}

	sess = &ds.UserSession{
		ID:        uuid.New(),
		UserID:    in.UserID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(app.Config().Session.DurationHours)),
	}

	err = s.db.CreateUserSession(ctx, sess)
	if err != nil {
		return
	}

	return
}

// CreateUserSessionInput ...
type CreateUserSessionInput struct {
	UserID int64
}

// Sanitize ...
func (in *CreateUserSessionInput) Sanitize() {}

// Validate ...
func (in *CreateUserSessionInput) Validate() error {
	return validateInput(createUserSessionInputRules, in)
}
