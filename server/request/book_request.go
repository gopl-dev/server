package request

// CreateBook ...
type CreateBook struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ReleaseDate string `json:"release_date"`
	AuthorName  string `json:"author_name"`
	AuthorLink  string `json:"author_link"`
	Homepage    string `json:"homepage"`
	CoverImage  string `json:"cover_image"`
}
