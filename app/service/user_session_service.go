package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

const (
	ctxUserSessionKey contextKey = "user_session"
)

func (s *Service) CreateUserSession(ctx context.Context, userID int64) (sess *ds.UserSession, err error) {
	sess = &ds.UserSession{
		ID:        uuid.New(),
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(app.Config().Session.DurationHours)),
	}

	err = s.db.CreateUserSession(ctx, sess)
	return
}

func (s *Service) FindUserSessionByID(ctx context.Context, id string) (sess *ds.UserSession, err error) {
	return s.db.FindUserSessionByID(ctx, id)
}

func (s *Service) ProlongUserSession(ctx context.Context, id uuid.UUID) (err error) {
	err = s.db.ProlongUserSession(ctx, id)
	return
}

func (s *Service) DeleteUserSession(ctx context.Context, id uuid.UUID) (err error) {
	err = s.db.DeleteUserSession(ctx, id)
	return
}

func (s *Service) UserSessionToContext(ctx context.Context, session *ds.UserSession) context.Context {
	return context.WithValue(ctx, ctxUserSessionKey, session)
}

func (s *Service) UserSessionFromContext(ctx context.Context) *ds.UserSession {
	if v := ctx.Value(ctxUserSessionKey); v != nil {
		return v.(*ds.UserSession)
	}

	return nil
}
