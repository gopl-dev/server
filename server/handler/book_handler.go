//nolint:all
package handler

import (
	"net/http"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/service"
)

type FilterBookRequest struct {
	Limit  int    `json:"limit" q:"limit"`
	Offset int    `json:"offset" q:"offset"`
	Name   string `json:"name" q:"name"`
}

func (r *FilterBookRequest) ToParams() service.FilterBooksParams {
	return service.FilterBooksParams{}
}

type FilterBookResponse struct {
	Count int       `json:"count"`
	Data  []ds.Book `json:"data"`
}

func (h *Handler) FilterBooks(w http.ResponseWriter, r *http.Request) {
	var req FilterBookRequest

	res := handleQueryRequest(w, r, &req)
	if res.Aborted() {
		return
	}

	books, count, err := h.service.FilterBooks(req.ToParams())
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonOK(FilterBookResponse{
		Count: count,
		Data:  books,
	})
}

func (h *Handler) GetBookByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("book_id")
	jsonOK(w, map[string]string{"id": id})
}
