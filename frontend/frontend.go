package frontend

import (
	"embed"

	"github.com/gopl-dev/server/app/ds"
)

//go:embed assets
var AssetsFs embed.FS

type User struct {
	ID       int64
	Username string
}

func NewUser(u *ds.User) *User {
	if u == nil {
		return nil
	}

	return &User{
		ID:       u.ID,
		Username: u.Username,
	}
}
