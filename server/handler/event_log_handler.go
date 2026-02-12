package handler

import (
	"net/http"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
)

// FilterEventLogsView renders the logs listing page with filtering UI.
func (h *Handler) FilterEventLogsView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "FilterEventLogsView")
	defer span.End()

	RenderDefaultLayout(ctx, w, layout.Data{
		Title: "Activity Log",
		Body:  page.FilterEventLogsPage(),
	})
}

// FilterEventLogs handles API requests for retrieving a filtered list of logs.
//
//	@ID			FilterEventLogs
//	@Summary	Get activity log
//	@Tags		event-logs
//	@Accept		json
//	@Produce	json
//	@Param		params	query		request.FilterEventLogs			false	"Query parameters"
//	@Success	200		{object}	response.FilterEventLogs
//	@Failure	400		{object}	Error
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/event-logs/ [get]
//	@Security	ApiKeyAuth
func (h *Handler) FilterEventLogs(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "FilterEventLogs")
	defer span.End()

	var req request.FilterEventLogs
	bindQuery(r, &req)

	logs, count, err := h.service.FilterEventLogs(ctx, ds.EventLogsFilter{
		Page:       req.Page,
		PerPage:    req.PerPage,
		OnlyPublic: true,
		WithCount:  true,
	})
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonOK(w, response.NewFilterEventLog(logs, count))
}

// EventLogChanges return changes of event log
//
//	@ID			EventLogChanges
//	@Summary	Get changes in an event log
//	@Tags		event-logs
//	@Accept		json
//	@Produce	json
//	@Param		id	path		string	true	"Event log ID"
//	@Success	200		{object}	response.EventLogChanges
//	@Failure	400		{object}	Error
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/event-logs/ [get]
//	@Security	ApiKeyAuth
func (h *Handler) EventLogChanges(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "EventLogChanges")
	defer span.End()

	id, err := idFromPath(r)
	if err != nil {
		Abort(w, r, err)
		return
	}

	changes, err := h.service.EventLogChanges(ctx, id)
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonOK(w, response.EventLogChanges{
		Changes: changes,
	})
}
