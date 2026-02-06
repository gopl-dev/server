package ds

import (
	"time"
)

// EntityChangeStatus represents the lifecycle state of an entity change request.
type EntityChangeStatus string

const (
	// EntityChangePending indicates a change request awaiting review.
	EntityChangePending EntityChangeStatus = "pending"

	// EntityChangeCommitted indicates an approved and applied change request.
	EntityChangeCommitted EntityChangeStatus = "committed"

	// EntityChangeRejected indicates a change request that was reviewed and rejected.
	EntityChangeRejected EntityChangeStatus = "rejected"
)

// EntityChangeRequest describes a user's requested changes to an entity.
type EntityChangeRequest struct {
	ID       ID                 `json:"id"`
	EntityID ID                 `json:"entity_id,omitzero"`
	UserID   ID                 `json:"user_id,omitzero"`
	Status   EntityChangeStatus `json:"status"`
	Diff     map[string]any     `json:"diff"`
	// State      map[string]any     `json:"state"`
	Message    string     `json:"message,omitempty"`
	Revision   int        `json:"revision"`
	ReviewerID *ID        `json:"reviewer_id,omitempty"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	ReviewNote string     `json:"review_note,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`

	// props needed to be returned from filter action
	Username       string     `json:"username"`
	EntityType     EntityType `json:"entity_type"`
	EntityTitle    string     `json:"entity_title"`
	EntityPublicID string     `json:"entity_public_id"`
}

// ChangeRequestsFilter is used to filter and paginate change requests.
type ChangeRequestsFilter struct {
	Page      int
	PerPage   int
	Status    EntityChangeStatus
	WithCount bool
}
