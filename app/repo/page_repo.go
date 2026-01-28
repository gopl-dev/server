package repo

import (
	"github.com/gopl-dev/server/app"
)

var (
	// ErrPageNotFound is a sentinel error returned when page not found.
	ErrPageNotFound = app.ErrNotFound("page not found")
)
