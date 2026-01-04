package ds

import z "github.com/Oudwins/zog"

// Book defines the data structure for a book.
type Book struct {
	Entity

	Description string `json:"description"`
	AuthorName  string `json:"author_name"`
	AuthorLink  string `json:"author_link"`
	Homepage    string `json:"homepage"`
	ReleaseDate string `json:"release_date"`
	CoverImage  string `json:"cover_image"`
}

// CreateRules provides the validation map used when saving a new book.
func (b *Book) CreateRules() z.Shape {
	return z.Shape{
		"Title":       z.String().Trim().Required(),
		"Description": z.String().Trim().Required(),
		"AuthorName":  z.String().Trim().Required(),
		"AuthorLink":  z.String().Trim().URL(),
		"Homepage":    z.String().Trim().URL().Required(),
		"ReleaseDate": z.String().Trim().Required(),
		"CoverImage":  z.String().Trim().URL().Required(),
	}
}

// UpdateRules provides the validation map used when editing an existing book.
func (b *Book) UpdateRules() z.Shape {
	return b.CreateRules()
}
