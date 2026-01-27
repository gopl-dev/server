package ds

import "slices"

// EntityType defines the type of the entity content.
type EntityType string

const (
	// EntityTypeBook ...
	EntityTypeBook EntityType = "book"
	EntityTypePage EntityType = "page"
)

// EntityTypes ...
var EntityTypes = []EntityType{
	EntityTypeBook,
	EntityTypePage,
}

// Valid ...
func (t EntityType) Valid() bool {
	return slices.Contains(EntityTypes, t)
}
