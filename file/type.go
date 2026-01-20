package file

import (
	"path/filepath"
	"strings"
)

// Type represents a file category derived from its extension.
type Type string

const (
	// TypeOther is used when file type cannot be determined
	// or does not belong to any known category.
	TypeOther Type = "other"

	// TypeImage represents image files.
	TypeImage Type = "image"

	// TypeVideo represents video files.
	TypeVideo Type = "video"
)

// fileTypes defines supported file extensions grouped by file type.
var fileTypes = map[Type][]string{
	TypeImage: {"jpeg", "jpg", "png", "bmp", "gif", "svg"},
	TypeVideo: {"avi", "mov", "mpg", "mpeg", "webm", "mkv", "3gp", "mpe", "mp4"},
}

// fileTypesByExt maps a file extension (with leading dot, as returned by
// filepath.Ext) to its corresponding Type.
// It is initialized in init() for faster lookup.
var fileTypesByExt map[string]Type

func init() {
	// Build a reverse lookup map from file extension to file type.
	// Using a precomputed map avoids iterating over all types on each call.
	fileTypesByExt = map[string]Type{}
	for t, exts := range fileTypes {
		for _, ext := range exts {
			// Prepend dot to match filepath.Ext() format (e.g. ".jpg").
			fileTypesByExt["."+ext] = t
		}
	}
}

// ResolveFileType determines the file Type based on the file extension.
// The lookup is case-insensitive. If the extension is unknown or missing,
// TypeOther is returned.
func ResolveFileType(path string) Type {
	ext := strings.ToLower(filepath.Ext(path))
	if ft, ok := fileTypesByExt[ext]; ok {
		return ft
	}

	return TypeOther
}
