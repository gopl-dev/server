package ds

import (
	"time"
)

// ActionLogType is an enumeration representing the different categories of actions
// that are recorded in the system logs.
type ActionLogType int

// ActionLog represents a single recorded action or event within the system,
// used for auditing and tracing system activities.
type ActionLog struct {
	ID        int64
	UserID    int64
	Log       string
	CreatedAt time.Time
}
