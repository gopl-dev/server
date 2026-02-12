package ds

import (
	"maps"
	"slices"
	"sort"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds/prop"
	"github.com/lithammer/shortuuid"
)

// Entity represents the base metadata for any user-content in the system.
type Entity struct {
	ID            ID               `json:"id"`
	PublicID      string           `json:"public_id"`
	OwnerID       ID               `json:"-"`
	PreviewFileID ID               `json:"preview_file_id"`
	Type          EntityType       `json:"-"`
	Title         string           `json:"title"`
	SummaryRaw    string           `json:"-"`
	Summary       string           `json:"summary"`
	Status        EntityStatus     `json:"status"`
	Visibility    EntityVisibility `json:"visibility"`
	PublishedAt   *time.Time       `json:"published_at,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     *time.Time       `json:"updated_at,omitempty"`
	DeletedAt     *time.Time       `json:"-"`

	Topics []Topic `json:"topics"`
	Owner  *string `db:"owner" json:"owner,omitempty"`
}

// CreateRules returns the validation schema for creating a new entity.
func (e *Entity) CreateRules() z.Shape {
	return z.Shape{
		"ID":      IDInputRules,
		"OwnerID": IDInputRules,
		"Type": z.CustomFunc(func(val *EntityType, _ z.Ctx) bool {
			return val.Valid()
		}),
		"Visibility": z.CustomFunc(func(val *EntityVisibility, _ z.Ctx) bool {
			return val.Valid()
		}),
		"Status": z.CustomFunc(func(val *EntityStatus, _ z.Ctx) bool {
			return slices.Contains([]EntityStatus{
				EntityStatusUnderReview,
				EntityStatusApproved,
				// Rejected is not valid status
			}, *val)
		}),
		"PublicID": z.String().Trim().Required(),
	}
}

// Data returns the editable fields of the Book as a key-value map.
func (e *Entity) Data() map[string]any {
	topics := make([]string, len(e.Topics))
	for i, t := range e.Topics {
		topics[i] = t.PublicID
	}
	sort.Strings(topics)

	return map[string]any{
		"public_id":       e.PublicID,
		"preview_file_id": e.PreviewFileID,
		"title":           e.Title,
		"summary":         e.SummaryRaw,
		"topics":          topics,
	}
}

// PropertyType returns the property type for a given key.
func (e *Entity) PropertyType(key string) prop.Type {
	switch key {
	case "public_id":
		return prop.String
	case "title":
		return prop.String
	case "summary":
		return prop.Markdown
	case "topics":
		return prop.List
	}

	return prop.Unknown
}

// WithEntityData merges the provided data map with the entity's existing data
// and returns the combined result.
func (e *Entity) WithEntityData(data map[string]any) map[string]any {
	d := e.Data()
	maps.Copy(d, data)

	return d
}

// UpdateRules returns the validation schema for updating an existing entity.
func (e *Entity) UpdateRules() z.Shape {
	return e.CreateRules()
}

// SetPublicID ensures that the entity has a non-empty, human-readable PublicID.
// If it cannot be derived from the Title, it falls back to "{type}_{shortuuid}".
func (e *Entity) SetPublicID() {
	if strings.TrimSpace(e.PublicID) == "" {
		e.PublicID = app.Slug(e.Title)
	}

	if strings.TrimSpace(e.PublicID) == "" {
		e.PublicID = string(e.Type) + "_" + shortuuid.New()
	}
}

// ViewURL returns the public-facing URL path for viewing the entity.
func (e *Entity) ViewURL() string {
	switch e.Type {
	case EntityTypeBook:
		return "/books/" + e.PublicID
	case EntityTypePage:
		return "/" + e.PublicID
	}

	return "/"
}

// EntitiesFilter is used to filter entities.
type EntitiesFilter struct {
	Page           int
	PerPage        int
	WithCount      bool
	CreatedAt      *FilterDT
	DeletedAt      *FilterDT
	Deleted        bool
	Title          *FilterString
	Visibility     []EntityVisibility
	Status         []EntityStatus
	Topics         []string
	OrderBy        string
	OrderDirection string
}
