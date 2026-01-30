package request

import "github.com/gopl-dev/server/app/ds"

// FilterEntities defines common pagination and search parameters.
type FilterEntities struct {
	Page       int                   `json:"page" url:"page,omitempty"`
	PerPage    int                   `json:"per_page" url:"per_page,omitempty"`
	Status     []ds.EntityStatus     `json:"s" url:"s,omitempty"`
	Visibility []ds.EntityVisibility `json:"v" url:"v,omitempty"`
	Search     *string               `json:"search" url:"search,omitempty"`
}
