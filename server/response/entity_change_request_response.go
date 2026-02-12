package response

import (
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/service"
)

// FilterChangeRequests represents a paginated collection of change requests returned by a filter operation.
type FilterChangeRequests struct {
	Data  []ds.EntityChangeRequest `json:"data"`
	Count int                      `json:"count"`
}

// ChangeRequestDiff represents the differences in a change request.
type ChangeRequestDiff struct {
	Diff []service.ChangeDiff `json:"diff"`
}
