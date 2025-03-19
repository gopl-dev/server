package ds

import (
	"time"
)

type ActionLogType int

type ActionLog struct {
	ID        int64
	UserID    int64
	Log       string
	CreatedAt time.Time
}
