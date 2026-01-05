package ds

import (
	"context"
	"time"
)

const (
	userCtxKey ctxKey = "user"

	// CleanupDeletedUserAfter ...
	CleanupDeletedUserAfter = 30 * 24 * time.Hour
)

// User ...
type User struct {
	ID             ID         `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	EmailConfirmed bool       `json:"-"`
	Password       string     `json:"-"`
	CreatedAt      time.Time  `json:"-"`
	UpdatedAt      *time.Time `json:"-"`
	DeletedAt      *time.Time `json:"-"`

	// IsAdmin is true if the user's ID is in the admin list from the config file.
	// This field is set by the auth middleware.
	IsAdmin bool `json:"-"`
}

// UsersFilter is used to filter and paginate user queries.
type UsersFilter struct {
	Page           int
	PerPage        int
	WithCount      bool
	CreatedAt      *FilterDT
	DeletedAt      *FilterDT
	Deleted        bool
	OrderBy        string
	OrderDirection string
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
