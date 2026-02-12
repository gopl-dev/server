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
//	@Param		params	query		request.FilterChangeRequests			false	"Query parameters"
//	@Success	200		{object}	response.FilterChangeRequests
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

// GetChangeRequestDiff retrieves and returns the diff for a specific change request.
//
//	@ID			GetChangeRequestDiff
//	@Summary	Get change requests diff for review
//	@Tags		change-requests
//	@Accept		json
//	@Produce	json
//	@Param		id	path string	true "Change request ID"
//	@Success	200		{object}	response.ChangeRequestDiff
//	@Failure	400		{object}	Error
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/change-requests/{id}/diff/ [get]
//	@Security	ApiKeyAuth
func (h *Handler) GetChangeRequestDiff(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "GetChangeRequestDiff")
	defer span.End()

	id, err := idFromPath(r)
	if err != nil {
		Abort(w, r, err)
		return
	}

	diff, _, err := h.service.GetChangeRequestDiff(ctx, id)
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonOK(w, response.ChangeRequestDiff{
		Diff: diff,
	})
}

// ApplyChangeRequest applies a pending change request to the entity.
//
//	@ID			ApplyChangeRequest
//	@Summary	Apply a pending change request to the entity.
//	@Tags		change-requests
//	@Accept		json
//	@Produce	json
//	@Param		id	path		string	true	"Change request ID"
//	@Success	200		{object}	service.ChangeDiff
//	@Failure	400		{object}	Error
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/change-requests/{id}/diff/ [put]
//	@Security	ApiKeyAuth
func (h *Handler) ApplyChangeRequest(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ApplyChangeRequest")
	defer span.End()

	id, err := idFromPath(r)
	if err != nil {
		Abort(w, r, err)
		return
	}

	err = h.service.ApplyChangeRequest(ctx, id)
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonOK(w, response.Success)
}

// RejectChangeRequest rejects a pending change request with a review note.
//
//	@ID			RejectChangeRequest
//	@Summary	Reject a pending change request
//	@Tags		change-requests
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.RejectBook	true	"Request body"
//	@Param		id	path		string	true	"Change request ID"
//	@Success	200		{object}	service.ChangeDiff
//	@Failure	400		{object}	Error
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/change-requests/{id}/diff/ [put]
//	@Security	ApiKeyAuth
func (h *Handler) RejectChangeRequest(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RejectChangeRequest")
	defer span.End()

	var req request.RejectBook
	user, res := handleAuthorizedJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	id, err := idFromPath(r)
	if err != nil {
		Abort(w, r, err)
		return
	}

	err = h.service.RejectChangeRequest(ctx, id, user.ID, req.Note)
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonOK(w, response.Success)
}
