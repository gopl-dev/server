package handler

import (
	"net/http"

	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
)

// FilterTopics handles the API request for creating a new book.
//
//	@ID			FilterTopics
//	@Summary	Filter topics
//	@Tags		topics
//	@Accept		json
//	@Produce	json
//	@Param		params	query		request.FilterTopics			false	"Query parameters"
//	@Success	200		{object}	response.FilterTopics
//	@Failure	400		{object}	Error
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/topics/ [get]
//	@Security	ApiKeyAuth
func (h *Handler) FilterTopics(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "FilterBooks")
	defer span.End()

	var req request.FilterTopics
	bindQuery(r, &req)

	f := req.ToFilter()
	f.WithCount = true

	topics, count, err := h.service.FilterTopics(ctx, f)
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonOK(w, response.FilterTopics{
		Data:  topics,
		Count: count,
	})
}
