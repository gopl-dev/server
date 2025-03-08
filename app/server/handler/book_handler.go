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

func FilterBooks(w http.ResponseWriter, r *http.Request) {
	var req FilterBookRequest
	h := handleQueryRequest(w, r, &req)
	if h.Aborted() {
		return
	}

	books, count, err := service.FilterBooks(req.ToParams())
	if err != nil {
		h.Abort(err)
		return
	}

	h.jsonOK(FilterBookResponse{
		Count: count,
		Data:  books,
	})
}

func GetBookByID(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]string{"status": "ok"})
}
