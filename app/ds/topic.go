package ds

import "time"

// Topic represents a topic scoped to a specific entity type (e.g. "book").
// Topics are linked to entities through the entity_topics pivot table.
type Topic struct {
	ID          ID         `json:"id,omitzero"`
	Type        EntityType `json:"entity_type,omitempty"`
	PublicID    string     `json:"public_id,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at,omitzero"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// EntityTopic links an entity to a topic (many-to-many).
type EntityTopic struct {
	Topic

	EntityID ID `json:"-"`
	TopicID  ID `json:"-"`
}

// TopicsFilter is used to filter and paginate user queries.
type TopicsFilter struct {
	Page           int
	PerPage        int
	Type           EntityType
	WithCount      bool
	OrderBy        string
	OrderDirection string
}
