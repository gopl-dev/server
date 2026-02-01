package request

import "github.com/gopl-dev/server/app/ds"

// FilterTopics defines input parameters for filtering and paginating topics.
type FilterTopics struct {
	Page    int           `json:"page" url:"page,omitempty"`
	PerPage int           `json:"per_page" url:"per_page,omitempty"`
	Type    ds.EntityType `json:"type" url:"type,omitempty"`
}

// ToFilter converts FilterTopics into a ds.TopicsFilter.
func (f FilterTopics) ToFilter() ds.TopicsFilter {
	return ds.TopicsFilter{
		Page:    f.Page,
		PerPage: f.PerPage,
		Type:    f.Type,
	}
}
