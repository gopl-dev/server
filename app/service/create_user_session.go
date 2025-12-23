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

	err = ValidateCreateUserSessionInput(userID)
	if err != nil {
		return
	}

	sess = &ds.UserSession{
		ID:        uuid.New(),
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(app.Config().Session.DurationHours)),
	}

	err = s.db.CreateUserSession(ctx, sess)
	return
}

// ValidateCreateUserSessionInput ...
func ValidateCreateUserSessionInput(userID int64) (err error) {
	in := &CreateUserSessionInput{
		UserID: userID,
	}

	return validateInput(createUserSessionInputRules, in)
}

// CreateUserSessionInput ...
type CreateUserSessionInput struct {
	UserID int64
}
