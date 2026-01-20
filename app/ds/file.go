package ds

import (
	"time"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/file"
)

const (
	// DeleteTempFilesAfterDays defines how long temporary files
	// are kept before being eligible for deletion.
	DeleteTempFilesAfterDays = 5

	// CleanupDeletedFilesAfterDays defines how long soft-deleted files
	// are kept before permanent cleanup.
	CleanupDeletedFilesAfterDays = 5
)

// File represents a file uploaded to the system along with its metadata.
// Some fields are internal-only and intentionally excluded from JSON output.
type File struct {
	ID          ID          `json:"id"`
	OwnerID     ID          `json:"owner_id"`
	Name        string      `json:"name"`
	Path        string      `json:"-"`
	PreviewPath string      `json:"-"`
	Hash        string      `json:"-"`
	Type        file.Type   `json:"type"`
	MimeType    string      `json:"mime_type"`
	Purpose     FilePurpose `json:"purpose"`
	Size        int64       `json:"size"`
	CreatedAt   time.Time   `json:"created_at"`
	DeletedAt   *time.Time  `json:"-"`
	Temp        bool        `json:"-"`
}

// CreateRules returns the validation schema for creating a new File.
func (f *File) CreateRules() z.Shape {
	return z.Shape{
		"ID":      IDInputRules,
		"OwnerID": IDInputRules,
		"Size":    z.Int64().GT(0).Required(),
		"Name":    z.String().Trim().Required(),
		"Path":    z.String().Trim().Required(),
		"Hash":    z.String().Trim().Required(),
		"Type": z.CustomFunc(func(val *file.Type, _ z.Ctx) bool {
			if val == nil || !val.Valid() {
				return false
			}

			return true
		}, z.Message("Invalid type")),
		"MimeType": z.String().Trim().Required(),
		"Purpose": z.CustomFunc(func(val *FilePurpose, _ z.Ctx) bool {
			if val == nil || !val.Valid() {
				return false
			}

			return true
		}, z.Message("Invalid purpose")),
	}
}

// IsOwner reports whether the file belongs to the given owner.
func (f *File) IsOwner(ownerID ID) bool {
	return f.OwnerID == ownerID
}

// IsBookCover reports whether the file is an image intended to be used
// as a book cover.
func (f *File) IsBookCover() bool {
	return f.Type == file.TypeImage && f.Purpose == FilePurposeBookCover
}

// FilesFilter is used to filter, sort, and paginate file queries.
type FilesFilter struct {
	Page           int
	PerPage        int
	WithCount      bool
	CreatedAt      *FilterDT
	DeletedAt      *FilterDT
	Deleted        bool
	OrderBy        string
	OrderDirection string
}
