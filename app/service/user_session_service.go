package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

// CreateUserSession creates a new user session object.
func (s *Service) CreateUserSession(ctx context.Context, userID int64) (sess *ds.UserSession, err error) {
	ctx, span := s.tracer.Start(ctx, "CreateUserSession")
	defer span.End()

	sess = &ds.UserSession{
		ID:        uuid.New(),
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(app.Config().Session.DurationHours)),
	}

	err = s.db.CreateUserSession(ctx, sess)
	return
}

// FindUserSessionByID retrieves a user session from the database using its ID.
func (s *Service) FindUserSessionByID(ctx context.Context, id string) (sess *ds.UserSession, err error) {
	ctx, span := s.tracer.Start(ctx, "FindUserSessionByID")
	defer span.End()

	return s.db.FindUserSessionByID(ctx, id)
}

// ProlongUserSession updates the expiration time of an existing user session in the database.
func (s *Service) ProlongUserSession(ctx context.Context, id uuid.UUID) (err error) {
	ctx, span := s.tracer.Start(ctx, "ProlongUserSession")
	defer span.End()

	err = s.db.ProlongUserSession(ctx, id)
	return
}

// DeleteUserSession removes a user session record from the database using its ID.
func (s *Service) DeleteUserSession(ctx context.Context, id uuid.UUID) (err error) {
	ctx, span := s.tracer.Start(ctx, "DeleteUserSession")
	defer span.End()

	err = s.db.DeleteUserSession(ctx, id)
	return
}
