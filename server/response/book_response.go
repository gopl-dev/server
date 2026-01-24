package response

import "github.com/gopl-dev/server/app/ds"

// FilterBooks ...
type FilterBooks struct {
	Data  []ds.Book `json:"data"`
	Count int       `json:"count"`
}
