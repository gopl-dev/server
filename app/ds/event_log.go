package ds

import (
	"strings"
	"time"

	"github.com/gopl-dev/server/app"
)

// EventLog represents a single event entry used for activity feeds,
// audits, and system transparency.
type EventLog struct {
	ID           ID      `json:"id"`
	UserID       *ID     `json:"user_id"`
	UserUsername *string `json:"-"`

	Type EventLogType `json:"action_type"`

	EntityID       *ID         `json:"entity_id"`
	EntityType     *EntityType `json:"-"`
	EntityTitle    *string     `json:"-"`
	EntityPublicID *string     `json:"-"`
	EntityChangeID *ID         `json:"entity_change_id"`

	// Message is an optional pre-rendered text for the event.
	// When set, it takes precedence over automatically generated messages.
	// This is primarily used for system events.
	Message string `json:"message"`

	// Meta holds custom data relevant to log
	Meta map[string]any `json:"-"`

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

	writeTitle := func(t string) {
		b.WriteString(` "`)
		if l.EntityPublicID != nil {
			b.WriteString(`<a href="/`)
			if *l.EntityType == EntityTypeBook {
				b.WriteString("books/")
			}
			b.WriteString(*l.EntityPublicID)
			b.WriteString(`/" class="link">`)
		}
		b.WriteString(t)
		if l.EntityPublicID != nil {
			b.WriteString(`</a>`)
		}
		b.WriteString(`"`)
	}

	// Actor: username
	if l.UserUsername != nil {
		b.WriteString(*l.UserUsername)
		b.WriteString(" ")
	}

	// Action verb (joined, created, updated, ...)
	b.WriteString(l.Type.Verb())

	// Entity type and title
	if l.EntityType != nil {
		b.WriteString(" ")
		b.WriteString(string(*l.EntityType))

		title := l.EntityTitle
		if storedTitle, ok := l.Meta["entity_title"]; ok {
			title = new(app.String(storedTitle))
		}

		if title != nil {
			writeTitle(*title)
			if l.Type == EventLogEntityRenamed {
				b.WriteString(" to ")
			}

			if renamedTo, ok := l.Meta["new_title"]; ok {
				writeTitle(app.String(renamedTo))
			}
		}
	}

	return b.String()
}

// EventLogsFilter is used to filter and paginate event logs.
type EventLogsFilter struct {
	Page       int
	PerPage    int
	OnlyPublic bool
	WithCount  bool
}
