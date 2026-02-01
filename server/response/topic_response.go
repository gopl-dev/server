package response

import "github.com/gopl-dev/server/app/ds"

// FilterTopics represents a paginated collection of topics returned by a filter operation.
type FilterTopics struct {
	Data  []ds.Topic `json:"data"`
	Count int        `json:"count"`
}
