package ds

import (
	"time"
)

// Action represents the type of operation performed on an entity.
type Action string

const (
	// ActionCreate indicates the initial creation of an entity.
	ActionCreate Action = "create"

	// ActionUpdate indicates a modification to an existing entity.
	ActionUpdate Action = "update"

	// ActionPublish indicates the entity was made publicly visible.
	ActionPublish Action = "publish"

	// ActionDelete indicates the entity was moved to a deleted state.
	ActionDelete Action = "delete"

	// ActionRestore indicates a previously deleted entity was recovered.
	ActionRestore Action = "restore"
)

// EntityChangeLog records a history of actions performed on an entity for auditing purposes.
type EntityChangeLog struct {
	ID       ID     `json:"id"`
	EntityID ID     `json:"entity_id"`
	UserID   ID     `json:"user_id"`
	Action   Action `json:"action"`
	// Rendered action for display purpose,
	// like "Book created", "book deleted"
	// if few changes were made, will to be explicit:
	// "Description of book updated", "Name and author of book was updated"
	Name      string         `json:"name"`
	Diff      map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}
