package ds

import (
	"time"
)

type AuthToken struct {
	ID         int64
	UserID     int64
	User       *User
	ClientName string
	ClientIP   string
	UserAgent  string
	Token      string
	CreatedAt  time.Time
	ExpiresAt  time.Time
}
