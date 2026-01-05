package ds

import "slices"

// EntityType defines the type of the entity content.
type EntityType string

const (
	// EntityTypeBook ...
	EntityTypeBook EntityType = "book"
)

// EntityTypes ...
var EntityTypes = []EntityType{
	EntityTypeBook,
}

// Valid ...
func (t EntityType) Valid() bool {
	return slices.Contains(EntityTypes, t)
}
