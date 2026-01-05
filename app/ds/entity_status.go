package ds

// EntityStatus defines the moderation state of an entity.
type EntityStatus string

const (
	// EntityStatusUnderReview means the entity is awaiting moderation.
	EntityStatusUnderReview EntityStatus = "review"

	// EntityStatusApproved means the entity is live and approved.
	EntityStatusApproved EntityStatus = "approved"

	// EntityStatusRejected means the entity failed moderation.
	EntityStatusRejected EntityStatus = "rejected"
)
