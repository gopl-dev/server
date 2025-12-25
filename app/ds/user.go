package ds

import (
	"context"
	"time"
)

const (
	userCtxKey ctxKey = "user"
)

// User ...
type User struct {
	ID             int64      `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	EmailConfirmed bool       `json:"-"`
	Password       string     `json:"-"`
	CreatedAt      time.Time  `json:"-"`
	UpdatedAt      *time.Time `json:"-"`
	DeletedAt      *time.Time `json:"-"`
}

// Deleted ...
func (u *User) Deleted() bool {
	return u.DeletedAt != nil
}

// ToContext adds the given user object to the provided context.
func (u *User) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

// UserFromContext attempts to retrieve user object from the context.
func UserFromContext(ctx context.Context) *User {
	if v := ctx.Value(userCtxKey); v != nil {
		if user, ok := v.(*User); ok {
			return user
		}
	}

	return nil
}
