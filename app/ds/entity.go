package ds

import (
	"slices"
	"time"

	z "github.com/Oudwins/zog"
)

// EntityType defines the category of the content and its relation to a specific data table.
type EntityType int

const (
	// EntityTypeBook represents a book resource.
	EntityTypeBook EntityType = iota + 1
)

var entityTypes = []EntityType{
	EntityTypeBook,
}

// EntityVisibility controls how the entity is accessed in listings and search results.
type EntityVisibility int

const (
	// EntityVisibilityDraft means the entity is visible only to the owner.
	EntityVisibilityDraft EntityVisibility = iota

	// EntityVisibilityPublic means the entity is visible to everyone and indexed.
	EntityVisibilityPublic

	// EntityVisibilityPrivate means the entity is restricted to the owner and collaborators.
	EntityVisibilityPrivate

	// EntityVisibilityUnlisted means the entity is accessible only via direct link.
	EntityVisibilityUnlisted
)

var entityVisibilities = []EntityVisibility{
	EntityVisibilityDraft,
	EntityVisibilityPublic,
	EntityVisibilityPrivate,
	EntityVisibilityUnlisted,
}

// EntityStatus defines the moderation state of an entity.
type EntityStatus int

const (
	// EntityStatusUnderReview means the entity is awaiting moderation.
	EntityStatusUnderReview EntityStatus = iota

	// EntityStatusPublished means the entity is live and approved.
	EntityStatusPublished

	// EntityStatusRejected means the entity failed moderation.
	EntityStatusRejected
)

// Entity represents the base metadata for any user-content in the system.
type Entity struct {
	ID          ID               `json:"id"`
	OwnerID     ID               `json:"owner_id"`
	Type        EntityType       `json:"-"`
	URLName     string           `json:"url_name"`
	Title       string           `json:"title"`
	Visibility  EntityVisibility `json:"visibility"`
	Status      EntityStatus     `json:"status"`
	PublishedAt *time.Time       `json:"published_at,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   *time.Time       `json:"updated_at,omitempty"`
	DeletedAt   *time.Time       `json:"-"`
}

// CreateRules returns the validation schema for creating a new entity.
func (e *Entity) CreateRules() z.Shape {
	return z.Shape{
		"ID":      IDInputRules,
		"OwnerID": IDInputRules,
		"Type": z.CustomFunc(func(val *EntityType, _ z.Ctx) bool {
			return slices.Contains(entityTypes, *val)
		}),
		"Visibility": z.CustomFunc(func(val *EntityVisibility, _ z.Ctx) bool {
			return slices.Contains(entityVisibilities, *val)
		}),
		"Status": z.CustomFunc(func(val *EntityStatus, _ z.Ctx) bool {
			return slices.Contains([]EntityStatus{
				EntityStatusUnderReview,
				EntityStatusPublished,
			}, *val)
		}),
		"URLName": z.String().Trim().Required(),
	}
}

// UpdateRules returns the validation schema for updating an existing entity.
func (e *Entity) UpdateRules() z.Shape {
	return e.CreateRules()
}
