package ds

import (
	"time"
)

type ContentImportStatus int

const (
	ImportSuccess ContentImportStatus = iota + 1
	ImportFailed
)

type ContentImportLog struct {
	ID        int64               `json:"id"`
	Path      string              `json:"path"`
	Branch    string              `json:"default_branch"`
	Status    ContentImportStatus `json:"import_status"`
	Log       string              `json:"import_log"`
	CreatedAt time.Time           `json:"created_at"`
}
