package response

import "github.com/gopl-dev/server/app/ds"

// FilterBooks represents a paginated collection of books returned by a filter operation.
type FilterBooks struct {
	Data  []ds.Book `json:"data"`
	Count int       `json:"count"`
}

// UpdateRevision defines the response payload for an update request.
//
// Revision is the current change-request revision for the entity:
//   - 0 means the update was applied immediately (if the user is authorized to do so).
//   - 1..N means the update was saved as a change request (review required),
//     where N increases with each subsequent change request for the same entity
//     and resets to 1 once the changes are reviewed.
type UpdateRevision struct {
	Revision int `json:"revision"`
}
