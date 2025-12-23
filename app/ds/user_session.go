package ds

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	userSessionCtxKey ctxKey = "user_session"
)

// UserSession represents an active session for a logged-in user.
type UserSession struct {
	ID        uuid.UUID  `json:"id"`
	UserID    int64      `json:"user_id"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt *time.Time `json:"-"`
	ExpiresAt time.Time  `json:"-"`
}

// ToContext adds the given user session object to the provided context.
func (s *UserSession) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, userSessionCtxKey, s)
}

// UserSessionFromContext attempts to retrieve the user session object from the context.
func UserSessionFromContext(ctx context.Context) *UserSession {
	if v := ctx.Value(userSessionCtxKey); v != nil {
		if session, ok := v.(*UserSession); ok {
			return session
		}
	}

	return nil
}
