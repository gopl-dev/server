package request

// FilterEntities defines common pagination and search parameters.
type FilterEntities struct {
	Page    int     `json:"page" url:"page,omitempty"`
	PerPage int     `json:"per_page" url:"per_page,omitempty"`
	Search  *string `json:"search" url:"search,omitempty"`
}
