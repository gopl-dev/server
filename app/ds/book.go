package ds

import (
	"context"

	z "github.com/Oudwins/zog"
)

var bookCtxKey ctxKey = "book"

// Book defines the data structure for a book.
type Book struct {
	*Entity

	CoverFileID ID     `json:"cover_file_id"`
	AuthorName  string `json:"author_name"`
	AuthorLink  string `json:"author_link"`
	Homepage    string `json:"homepage"`
	ReleaseDate string `json:"release_date"`
}

// Data returns the editable fields of the Book as a key-value map.
func (b *Book) Data() map[string]any {
	return map[string]any{
		"title":         b.Title,
		"cover_file_id": b.CoverFileID,
		"description":   b.Description,
		"author_name":   b.AuthorName,
		"author_link":   b.AuthorLink,
		"homepage":      b.Homepage,
		"release_date":  b.ReleaseDate,
		"topics":        b.Topics,
	}
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
	}
}

// UpdateRules provides the validation map used when editing an existing book.
func (b *Book) UpdateRules() z.Shape {
	return b.CreateRules()
}

// ToContext adds the given book object to the provided context.
func (b *Book) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, bookCtxKey, b)
}

// BookFromContext attempts to retrieve book object from the context.
func BookFromContext(ctx context.Context) *Book {
	if v := ctx.Value(bookCtxKey); v != nil {
		if book, ok := v.(*Book); ok {
			return book
		}
	}

	return nil
}

// BooksFilter is used to filter and paginate user queries.
type BooksFilter struct {
	EntitiesFilter
}
