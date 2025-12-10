package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

const (
	// ctxUserSessionKey is the unique context key used to store the active user session.
	ctxUserSessionKey contextKey = "user_session"
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

// UserSessionToContext adds the given user session object to the provided context.
func (s *Service) UserSessionToContext(ctx context.Context, session *ds.UserSession) context.Context {
	return context.WithValue(ctx, ctxUserSessionKey, session)
}

// UserSessionFromContext attempts to retrieve the user session object from the context.
func (s *Service) UserSessionFromContext(ctx context.Context) *ds.UserSession {
	if v := ctx.Value(ctxUserSessionKey); v != nil {
		// Safe type assertion to prevent panics
		if session, ok := v.(*ds.UserSession); ok {
			return session
		}
	}

	return nil
}
