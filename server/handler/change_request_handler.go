package handler

import (
	"net/http"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
)

// FilterChangeRequests handles API requests for retrieving a filtered list of change requests.
//
//	@ID			FilterChangeRequests
//	@Summary	Get change requests
//	@Tags		change-requests
//	@Accept		json
//	@Produce	json
//	@Param		params	query		request.FilterEventLogs			false	"Query parameters"
//	@Success	200		{object}	response.FilterEventLogs
//	@Failure	400		{object}	Error
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/change-requests/ [get]
//	@Security	ApiKeyAuth
func (h *Handler) FilterChangeRequests(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "FilterChangeRequests")
	defer span.End()

	var req request.FilterChangeRequests
	bindQuery(r, &req)

	data, count, err := h.service.FilterChangeRequests(ctx, ds.ChangeRequestsFilter{
		Page:      req.Page,
		PerPage:   req.PerPage,
		Status:    req.Status,
		WithCount: true,
	})
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonOK(w, response.FilterChangeRequests{
		Data:  data,
		Count: count,
	})
}

// GetChangeRequestReviewDiff handles API requests for retrieving a filtered list of change requests.
//
//	@ID			GetChangeRequestReviewDiff
//	@Summary	Get change requests diff for review
//	@Tags		change-requests
//	@Accept		json
//	@Produce	json
//	@Param		id	path		string	true	"Change request ID"
//	@Success	200		{object}	service.ChangeDiff
//	@Failure	400		{object}	Error
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/change-requests/{id}/diff/ [get]
//	@Security	ApiKeyAuth

func (h *Handler) GetChangeRequestReviewDiff(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "GetChangeRequestReviewDiff")
	defer span.End()

	id, err := idFromPath(r)
	if err != nil {
		Abort(w, r, err)
	}

	diff, err := h.service.GetChangeRequestReviewDiff(ctx, id)
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonOK(w, diff)
}
