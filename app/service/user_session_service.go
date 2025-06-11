package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
)

const (
	ctxUserSessionKey contextKey = "user_session"
)

func CreateUserSession(ctx context.Context, userID int64) (sess *ds.UserSession, err error) {
	sess = &ds.UserSession{
		ID:        uuid.New(),
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(app.Config().Session.DurationHours)),
	}

	err = repo.CreateUserSession(ctx, sess)
	return
}

func FindUserSessionByID(ctx context.Context, id string) (sess *ds.UserSession, err error) {
	return repo.FindUserSessionByID(ctx, id)
}

func ProlongUserSession(ctx context.Context, id uuid.UUID) (err error) {
	err = repo.ProlongUserSession(ctx, id)
	return
}

func DeleteUserSession(ctx context.Context, id uuid.UUID) (err error) {
	err = repo.DeleteUserSession(ctx, id)
	return
}

func UserSessionToContext(ctx context.Context, s *ds.UserSession) context.Context {
	return context.WithValue(ctx, ctxUserSessionKey, s)
}

func UserSessionFromContext(ctx context.Context) *ds.UserSession {
	if v := ctx.Value(ctxUserSessionKey); v != nil {
		return v.(*ds.UserSession)
	}

	return nil
}
