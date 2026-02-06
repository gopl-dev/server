package response

import "github.com/gopl-dev/server/app/ds"

// FilterBooks represents a paginated collection of books returned by a filter operation.
type FilterBooks struct {
	Data  []ds.Book `json:"data"`
	Count int       `json:"count"`
}
