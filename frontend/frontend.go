// Package frontend handles the application's user-facing elements, such as
// serving static files and rendering HTML templates
package frontend

import (
	"embed"

	"github.com/gopl-dev/server/app/ds"
)

// AssetsFs holds the content of the embedded FS for application assets
//
//go:embed assets
var AssetsFs embed.FS

// User is representation of a user.
type User struct {
	ID       ds.ID
	Username string
	IsAdmin  bool
}

// NewUser creates new User instance.
func NewUser(u *ds.User) *User {
	if u == nil {
		return nil
	}

	return &User{
		ID:       u.ID,
		Username: u.Username,
		IsAdmin:  u.IsAdmin,
	}
}
