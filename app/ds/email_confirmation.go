package ds

import (
	"time"
)

type EmailConfirmation struct {
	ID          int64
	UserID      int64
	Code        string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	ConfirmedAt *time.Time
}

func (c *EmailConfirmation) Invalid() bool {
	return c.ExpiresAt.Before(time.Now())
}
