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
	CoverFileID ds.ID  `json:"cover_file_id,omitzero"`

	Visibility ds.EntityVisibility `json:"visibility"`
}

type FilterEntities struct {
	Page    int     `json:"page" q:"page"`
	PerPage int     `json:"per_page" q:"per_page"`
	Search  *string `json:"search" q:"search"`
}

type FilterBooks struct {
	FilterEntities
}
