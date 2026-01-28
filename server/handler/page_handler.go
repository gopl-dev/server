package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
)

// RenderPageOrNotFound ...
func (h *Handler) RenderPageOrNotFound(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "GetBookView")
	defer span.End()

	id := strings.TrimPrefix(r.RequestURI, "/")
	id = strings.TrimSuffix(id, "/")

	p, err := h.service.GetPageByPublicID(ctx, id)
	if errors.Is(err, repo.ErrEntityNotFound) {
		RenderDefaultLayout(ctx, w, layout.Data{
			Title: "404 Not Found",
			Body:  page.Err404("This page does not exist."),
		})
		return
	}
	if err != nil {
		Abort(w, r, err)
		return
	}

	RenderDefaultLayout(ctx, w, layout.Data{
		Title: p.Title,
		Body:  page.ViewPage(id, p.Title, p.Description),
	})
}

// CreatePage handles the API request for creating a new page.
// TODO add openapi specs when this endpoint becomes public.
func (h *Handler) CreatePage(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "CreatePage")
	defer span.End()

	var req request.CreatePage
	user, res := handleAuthorizedJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	if !user.IsAdmin {
		Abort(w, r, app.ErrBadRequest("page creation is not available for now"))
		return
	}

	p := req.ToPage()
	p.OwnerID = user.ID

	err := h.service.CreatePage(ctx, p)
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonCreated(p)
}

// UpdatePage handles the API request for updating page.
//
//	@ID			UpdatePage
//	@Summary	Update page
//	@Tags		pages
//	@Accept		json
//	@Produce	json
//	@Param		id	path		string	true	"Page ID"
//	@Param		request	body		request.UpdatePage	true	"Request body"
//	@Success	200		{object}	response.UpdateRevision
//	@Failure	400		{object}	Error
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/pages/{id}/ [put]
//	@Security	ApiKeyAuth
func (h *Handler) UpdatePage(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "UpdatePage")
	defer span.End()

	var req request.UpdatePage
	_, res := handleAuthorizedJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	id := r.PathValue("id")
	revision, err := h.service.UpdatePage(ctx, id, req.ToPage())
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonOK(response.UpdateRevision{
		Revision: revision,
	})
}

// GetPageEditState return state of page changes for current user
//
//	@ID			GetPageEditState
//	@Summary	Get page for editing
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id	path		string	true	"Page ID"
//	@Success	201		{object}	service.EntityChange
//	@Failure	400		{object}	Error
//	@Failure	401		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/pages/{id}/edit/ [get]
//	@Security	ApiKeyAuth
func (h *Handler) GetPageEditState(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "GetPageEditState")
	defer span.End()

	p := ds.PageFromContext(ctx)
	if p == nil {
		Abort(w, r, app.ErrBadRequest("page is missing from context"))
		return
	}

	state, err := h.service.GetEntityChangeState(ctx, p.ID, p)
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonOK(w, state)
}

// EditPageView renders the static HTML page with the form for editing existing page.
func (h *Handler) EditPageView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "EditPageView")
	defer span.End()

	p := ds.PageFromContext(ctx)
	if p == nil {
		Abort(w, r, app.ErrBadRequest("page is missing from context"))
		return
	}

	RenderDefaultLayout(ctx, w, layout.Data{
		Title: "Edit page",
		Body:  page.EditPageForm(p.PublicID),
	})
}

// CreatePageView renders the static HTML page with the form for creating a new page.
func (h *Handler) CreatePageView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "CreatePageView")
	defer span.End()

	user := ds.UserFromContext(ctx)
	if user == nil {
		Abort(w, r, app.ErrUnauthorized())
		return
	}

	if !user.IsAdmin {
		Abort(w, r, app.ErrBadRequest("creating pages is not available for now"))
		return
	}

	RenderDefaultLayout(ctx, w, layout.Data{
		Title: "Create page",
		Body:  page.CreatePageForm(),
	})
}
