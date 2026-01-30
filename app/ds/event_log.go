package ds

import (
	"strings"
	"time"
)

// EventLog represents a single event entry used for activity feeds,
// audits, and system transparency.
type EventLog struct {
	ID     ID    `json:"id"`
	UserID *ID   `json:"user_id"`
	User   *User `json:"-"`

	Type EventLogType `json:"action_type"`

	EntityID       *ID     `json:"entity_id"`
	Entity         *Entity `json:"-"`
	EntityChangeID *ID     `json:"entity_change_id"`

	// Message is an optional pre-rendered text for the event.
	// When set, it takes precedence over automatically generated messages.
	// This is primarily used for system events.
	Message string `json:"message"`

	// Meta holds custom data relevant to log
	Meta map[string]any `json:"meta"`

	IsPublic  bool      `json:"is_public"`
	CreatedAt time.Time `json:"created_at"`
}

// RenderMessage builds a human-readable, public-facing description of the event.
//
// The message format is {user} {verb} [{entityType} "{title}"]
// e.g.
//   - ognev.dev created book "Hello world"
//   - ognev.dev joined
//
// If Message is explicitly set, it is returned as-is. This is primarily used
// for system events where the text cannot be reliably derived from structured data.
// Otherwise, the message is composed from User, Type, and Entity.
func (l EventLog) RenderMessage() string {
	if l.Message != "" {
		return l.Message
	}

	var b strings.Builder

	// Actor: username
	if l.User != nil && l.User.Username != "" {
		b.WriteString(l.User.Username)
		b.WriteString(" ")
	}

	// Action verb (joined, created, updated, ...)
	b.WriteString(l.Type.Verb())

	// Entity type and title
	if l.Entity != nil {
		if l.Entity.Type.Valid() {
			b.WriteString(" ")
			b.WriteString(string(l.Entity.Type))
		}

		if l.Entity.Title != "" {
			b.WriteString(` "`)
			b.WriteString(l.Entity.Title)
			b.WriteString(`"`)
		}
	}

	return b.String()
}
