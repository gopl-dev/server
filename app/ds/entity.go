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
			return val.Valid()
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
