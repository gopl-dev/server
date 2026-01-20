package ds

import "slices"

// FilePurpose describes the semantic purpose of a file in the system.
type FilePurpose string

const (
	// FilePurposeBookCover marks a file used as a book cover image.
	FilePurposeBookCover FilePurpose = "book-cover"
)

var filePurposes = []FilePurpose{
	FilePurposeBookCover,
}

// Valid ...
func (p FilePurpose) Valid() bool {
	return slices.Contains(filePurposes, p)
}
