package request

import "github.com/gopl-dev/server/app/ds"

// CreateBook ...
type CreateBook struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ReleaseDate string `json:"release_date"`
	AuthorName  string `json:"author_name"`
	AuthorLink  string `json:"author_link"`
	Homepage    string `json:"homepage"`
	CoverImage  string `json:"cover_image"`

	Visibility ds.EntityVisibility `json:"visibility"`
}
