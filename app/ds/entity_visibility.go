package ds

import "slices"

// EntityVisibility controls how the entity is accessed in listings and search results.
type EntityVisibility string

const (

	// EntityVisibilityPublic means the entity is visible to everyone and indexed.
	EntityVisibilityPublic EntityVisibility = "public"

	// EntityVisibilityPrivate means the entity is restricted to the owner and collaborators.
	EntityVisibilityPrivate EntityVisibility = "private"

	// EntityVisibilityUnlisted means the entity is accessible only via direct link.
	EntityVisibilityUnlisted EntityVisibility = "unlisted"
)

// EntityVisibilities ...
var EntityVisibilities = []EntityVisibility{
	EntityVisibilityPublic,
	EntityVisibilityPrivate,
	EntityVisibilityUnlisted,
}

// Valid ...
func (v EntityVisibility) Valid() bool {
	return slices.Contains(EntityVisibilities, v)
}
