package response

import "github.com/gopl-dev/server/app/ds"

// FilterChangeRequests represents a paginated collection of change requests returned by a filter operation.
type FilterChangeRequests struct {
	Data  []ds.EntityChangeRequest `json:"data"`
	Count int                      `json:"count"`
}
