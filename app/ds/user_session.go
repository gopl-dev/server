package ds

import (
	"time"

	"github.com/google/uuid"
)

// UserSession represents an active session for a logged-in user.
type UserSession struct {
	ID        uuid.UUID  `json:"id"`
	UserID    int64      `json:"user_id"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt *time.Time `json:"-"`
	ExpiresAt time.Time  `json:"-"`
}
