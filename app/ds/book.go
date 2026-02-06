package ds

import (
	"context"
	"sort"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
)

var bookCtxKey ctxKey = "book"

// Book defines the data structure for a book.
type Book struct {
	*Entity

	DescriptionRaw string       `json:"-"`
	Description    string       `json:"description"`
	CoverFileID    ID           `json:"cover_file_id"`
	Authors        []BookAuthor `json:"authors"`
	Homepage       string       `json:"homepage"`
	ReleaseDate    string       `json:"release_date"`
}

// Data returns the editable fields of the Book as a key-value map.
func (b *Book) Data() map[string]any {
	if b.Topics == nil {
		b.Topics = make([]Topic, 0)
	}

	if b.Authors == nil {
		b.Authors = make([]BookAuthor, 0)
	}

	topics := make([]string, len(b.Topics))
	for i, t := range b.Topics {
		topics[i] = t.PublicID
	}
	sort.Strings(topics)

	return map[string]any{
		"title":         b.Title,
		"cover_file_id": b.CoverFileID,
		"summary":       b.SummaryRaw,
		"description":   b.DescriptionRaw,
		"homepage":      b.Homepage,
		"release_date":  b.ReleaseDate,
		"topics":        topics,
		"authors":       b.Authors,
	}
}

// ReleaseDateLayouts defines the allowed date layouts for formatting
// book release dates.
var ReleaseDateLayouts = []string{
	"2006",
	"January 2006",
	"January 2, 2006",
}

// CreateRules provides the validation map used when saving a new book.
func (b *Book) CreateRules() z.Shape {
	return z.Shape{
		"Title":       z.String().Trim().Required(),
		"Description": z.String().Required(),
		"Homepage":    z.String().Trim().URL(),
		"ReleaseDate": z.CustomFunc(func(val *string, _ z.Ctx) bool {
			if val == nil || *val == "" {
				return false
			}

			for _, layout := range ReleaseDateLayouts {
				_, err := time.Parse(layout, *val)
				if err == nil {
					return true
				}
			}

			return false
		}, z.Message("format must one of: "+strings.Join(ReleaseDateLayouts, "; "))),
		"Authors": z.Slice(z.Struct(z.Shape{
			"Name": z.String().Trim().Required(),
			"Link": z.String().Trim().URL(),
		})).Min(1).Required(),
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

// BookAuthor represents an author of a book.
type BookAuthor struct {
	Name string `json:"name"`
	Link string `json:"link"`
}
