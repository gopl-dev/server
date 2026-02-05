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
//	@ID			UpdateRevision
//	@Summary	Update book
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id	path		string	true	"Book ID"
//	@Param		request	body		request.UpdateBook	true	"Request body"
//	@Success	200		{object}	response.UpdateRevision
//	@Failure	400		{object}	Error
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/books/{id}/ [put]
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

	revision, err := h.service.UpdateBook(ctx, id, req.ToBook())
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonOK(response.UpdateRevision{
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

// FilterBooks handles API requests for retrieving a filtered list of books.
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

	var isAdmin bool
	user := ds.UserFromContext(ctx)
	if user != nil {
		isAdmin = user.IsAdmin
	}

	filter := ds.EntitiesFilter{
		Page:           req.Page,
		PerPage:        req.PerPage,
		WithCount:      true,
		Status:         req.Status,
		Visibility:     req.Visibility,
		Topics:         req.Topics,
		OrderBy:        "created_at",
		OrderDirection: "desc",
	}

	if !isAdmin {
		filter.Status = []ds.EntityStatus{ds.EntityStatusApproved}
		filter.Visibility = []ds.EntityVisibility{ds.EntityVisibilityPublic}
	}

	books, count, err := h.service.FilterBooks(ctx, ds.BooksFilter{
		EntitiesFilter: filter,
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

// FilterBooksView renders the books listing page with filtering UI.
func (h *Handler) FilterBooksView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "FilterBooksView")
	defer span.End()

	RenderDefaultLayout(ctx, w, layout.Data{
		Title: "Books",
		Body:  page.FilterBooksPage(),
	})
}

// GetBookView renders a single book details page.
func (h *Handler) GetBookView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "GetBookView")
	defer span.End()

	book := ds.BookFromContext(ctx)
	if book == nil {
		Abort(w, r, app.ErrBadRequest("book is missing from context"))
		return
	}

	RenderDefaultLayout(ctx, w, layout.Data{
		Title: book.Title,
		Body:  page.ViewBookPage(ds.UserFromContext(ctx), book),
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
	ctx, span := h.tracer.Start(r.Context(), "GetBookEditState")
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

// ApproveNewBook approves new book
//
//	@ID			ApproveNewBook
//	@Summary	Approve new book
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id	path		string	true	"Book ID"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	Error
//	@Failure	401		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/books/{id}/approve/ [put]
//	@Security	ApiKeyAuth
func (h *Handler) ApproveNewBook(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ApproveNewBook")
	defer span.End()

	book := ds.BookFromContext(ctx)
	if book == nil {
		Abort(w, r, app.ErrBadRequest("book is missing from context"))
		return
	}

	err := h.service.ApproveNewBook(ctx, book)
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonOK(w, response.Success)
}

// RejectNewBook approves new book
//
//	@ID			RejectNewBook
//	@Summary	Reject new book
//	@Tags		books
//	@Accept		json
//	@Produce	json
//	@Param		id	path		string	true	"Book ID"
//	@Param		request	body		request.RejectBook	true	"Request body"
//	@Success	201		{object}	response.Status
//	@Failure	400		{object}	Error
//	@Failure	401		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/books/{id}/reject/ [put]
//	@Security	ApiKeyAuth
func (h *Handler) RejectNewBook(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RejectNewBook")
	defer span.End()

	var req request.RejectBook
	_, res := handleAuthorizedJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	book := ds.BookFromContext(ctx)
	if book == nil {
		Abort(w, r, app.ErrBadRequest("book is missing from context"))
		return
	}

	err := h.service.RejectNewBook(ctx, req.Note, book)
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonOK(w, response.Success)
}

// CreateBookView renders the static HTML page with the form for creating a new book.
func (h *Handler) CreateBookView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "CreateBookView")
	defer span.End()

	RenderDefaultLayout(ctx, w, layout.Data{
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

	RenderDefaultLayout(ctx, w, layout.Data{
		Title: "Edit book",
		Body:  page.EditBookForm(book.ID.String()),
	})
}
