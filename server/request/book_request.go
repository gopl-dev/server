package request

import (
	"strings"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

// CreateBook defines the request payload for creating a new book entity.
type CreateBook struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	ReleaseDate string          `json:"release_date"`
	Homepage    string          `json:"homepage"`
	CoverFileID ds.ID           `json:"cover_file_id,omitempty,omitzero"`
	Authors     []ds.BookAuthor `json:"authors"`
	Topics      []ds.ID         `json:"topics"`
}

// Sanitize normalizes and validates CreateBook request.
func (r *CreateBook) Sanitize() {
	authors := make([]ds.BookAuthor, 0)
	for _, a := range r.Authors {
		n := strings.TrimSpace(a.Name)
		if n == "" {
			continue
		}
		authors = append(authors, ds.BookAuthor{
			Name: n,
			Link: strings.TrimSpace(a.Link),
		})
	}

	r.Authors = authors
}

// ToBook converts the CreateBook request into a Book model.
func (r *CreateBook) ToBook() *ds.Book {
	topics := make([]ds.Topic, len(r.Topics))
	for i, id := range r.Topics {
		topics[i] = ds.Topic{ID: id}
	}

	return &ds.Book{
		Entity: &ds.Entity{
			ID:            ds.NewID(),
			OwnerID:       ds.NilID,
			PreviewFileID: r.CoverFileID,
			Type:          ds.EntityTypeBook,
			PublicID:      app.Slug(r.Title),
			Title:         r.Title,
			Description:   r.Description,
			Visibility:    ds.EntityVisibilityPublic,
			Status:        ds.EntityStatusUnderReview,
			Topics:        topics,
			PublishedAt:   nil,
			CreatedAt:     time.Now(),
			UpdatedAt:     nil,
			DeletedAt:     nil,
		},
		Authors:     r.Authors,
		Homepage:    r.Homepage,
		ReleaseDate: r.ReleaseDate,
		CoverFileID: r.CoverFileID,
	}
}

// UpdateBook defines the request payload for updating an existing book.
// It reuses CreateBook fields as the updatable subset.
type UpdateBook struct {
	CreateBook
}

// FilterBooks defines filtering options specific to books.
type FilterBooks struct {
	FilterEntities
}
