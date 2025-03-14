package ds

import (
	"time"
)

type EntityData map[string]any

type Entity struct {
	ID        int64      `json:"id"`
	Path      string     `json:"path"`
	Title     string     `json:"title"`
	Type      string     `json:"type" pg:",use_zero"`
	Data      EntityData `json:"data"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`

	Topics []Topic `json:"-" pg:"-"`

	// composite fields
	SourceURL  string `json:"source_url" pg:"-"`
	EditURL    string `json:"edit_url" pg:"-"`
	CommitsURL string `json:"commits_url" pg:"-"`
}
