package response

import "github.com/gopl-dev/server/app/ds"

// FilterBooks represents a paginated collection of books returned by a list or search operation.
type FilterBooks struct {
	Data  []ds.Book `json:"data"`  // Data is the list of books for the current page.
	Count int       `json:"count"` // Count is the total number of matching books.
}

// UpdateBook defines the payload used to update an existing book.
//
// Revision == 0 means there is no review revision created and the book
// is updated directly (when the user is authorized to do so).
type UpdateBook struct {
	Revision int `json:"revision"`
}
