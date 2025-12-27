package ds

import (
	"database/sql"
	"encoding/json"
	"time"

	useractivity "github.com/gopl-dev/server/app/ds/user_activity"
)

// UserActivityLog represents a single entry in the user_activity_logs table.
type UserActivityLog struct {
	ID         ID                `json:"id"`
	UserID     ID                `json:"user_id"`
	ActionType useractivity.Type `json:"action_type"`
	IsPublic   bool              `json:"is_public"`
	EntityType sql.NullString    `json:"entity_type"`
	EntityID   sql.NullInt64     `json:"entity_id"`
	Meta       json.RawMessage   `json:"meta,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}
