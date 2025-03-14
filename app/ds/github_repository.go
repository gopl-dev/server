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
	ID        int64
	Status    ContentImportStatus
	Log       string
	CreatedAt time.Time
}
