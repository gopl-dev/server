package handler

import (
	"net/http"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
)

// CreateBook handles the API request for creating a new book.
//
//	@ID			CreateBook
//	@Summary	Create a new book
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.CreateBook	true	"Request body"
//	@Success	201		{object}	ds.Book
//	@Failure	400		{object}	Error
//	@Failure	401		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/books/ [post]
//	@Security	ApiKeyAuth
func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "CreateBook")
	defer span.End()

	var req request.CreateBook
	user, res := handleAuthorizedJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	book := req.ToBook()
	book.OwnerID = user.ID

	err := h.service.CreateBook(ctx, book)
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonCreated(book)
}

// UpdateBook handles the API request for updating book.
//
//	@ID			UpdateBook
//	@Summary	Update book
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.UpdateBook	true	"Request body"
//	@Success	200		{object}	response.UpdateBook
//	@Failure	400		{object}	Error
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/books/ [put]
//	@Security	ApiKeyAuth
func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "UpdateBook")
	defer span.End()

	var req request.UpdateBook
	_, res := handleAuthorizedJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	id, err := idFromPath(r)
	if err != nil {
		Abort(w, r, err)
		return
	}

	book := req.ToBook()

	revision, err := h.service.UpdateBook(ctx, id, book)
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonOK(response.UpdateBook{
		Revision: revision,
	})
}

// GetBook handles the API request for creating a new book.
//
//	@ID			GetBook
//	@Summary	Get book
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id	path		string	true	"Book ID"
//	@Success	201		{object}	ds.Book
//	@Failure	400		{object}	Error
//	@Failure	401		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/books/{id}/ [get]
//	@Security	ApiKeyAuth
func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "CreateBook")
	defer span.End()

	book := ds.BookFromContext(ctx)
	if book == nil {
		Abort(w, r, app.ErrBadRequest("book is missing from context"))
		return
	}

	jsonOK(w, book)
}

// FilterBooks handles the API request for creating a new book.
//
//	@ID			FilterBooks
//	@Summary	Filter books
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		params	query		request.FilterBooks			false	"Query parameters"
//	@Success	200		{object}	response.FilterBooks
//	@Failure	400		{object}	Error
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/books/ [get]
//	@Security	ApiKeyAuth
func (h *Handler) FilterBooks(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "FilterBooks")
	defer span.End()

	var req request.FilterBooks
	bindQuery(r, &req)

	books, count, err := h.service.FilterBooks(ctx, ds.BooksFilter{
		EntitiesFilter: ds.EntitiesFilter{
			Page:           req.Page,
			PerPage:        req.PerPage,
			WithCount:      true,
			Status:         []ds.EntityStatus{ds.EntityStatusApproved},
			Visibility:     []ds.EntityVisibility{ds.EntityVisibilityPublic},
			OrderBy:        "created_at",
			OrderDirection: "desc",
		},
	})
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonOK(w, response.FilterBooks{
		Data:  books,
		Count: count,
	})
}

// FilterBooksView ...
func (h *Handler) FilterBooksView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "FilterBooksView")
	defer span.End()

	renderDefaultLayout(ctx, w, layout.Data{
		Title: "Books",
		Body:  page.FilterBooksPage(),
	})
}

// GetBookEditState return state of book changes for current user
//
//	@ID			GetBookEditState
//	@Summary	Get book for editing
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id	path		string	true	"Book ID"
//	@Success	201		{object}	service.EntityChange
//	@Failure	400		{object}	Error
//	@Failure	401		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/books/{id}/edit/ [get]
//	@Security	ApiKeyAuth
func (h *Handler) GetBookEditState(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "GetEntityChangeRequest")
	defer span.End()

	book := ds.BookFromContext(ctx)
	if book == nil {
		Abort(w, r, app.ErrBadRequest("book is missing from context"))
		return
	}

	state, err := h.service.GetEntityChangeState(ctx, book.ID, book)
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonOK(w, state)
}

// CreateBookView renders the static HTML page with the form for creating a new book.
func (h *Handler) CreateBookView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "CreateBookView")
	defer span.End()

	renderDefaultLayout(ctx, w, layout.Data{
		Title: "Add book",
		Body:  page.CreateBookForm(),
	})
}

// EditBookView renders the static HTML page with the form for editing existing book.
func (h *Handler) EditBookView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "EditBookView")
	defer span.End()

	book := ds.BookFromContext(ctx)
	if book == nil {
		Abort(w, r, app.ErrBadRequest("book is missing from context"))
		return
	}

	renderDefaultLayout(ctx, w, layout.Data{
		Title: "Edit book",
		Body:  page.EditBookForm(book.ID.String()),
	})
}
