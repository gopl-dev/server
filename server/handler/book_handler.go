package handler

import (
	"net/http"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
	"github.com/gopl-dev/server/server/request"
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

	book := &ds.Book{
		Entity: ds.Entity{
			ID:          ds.NewID(),
			OwnerID:     user.ID,
			Type:        ds.EntityTypeBook,
			URLName:     "",
			Title:       req.Title,
			Visibility:  ds.EntityVisibilityDraft,
			Status:      ds.EntityStatusUnderReview,
			PublishedAt: nil,
			CreatedAt:   time.Now(),
			UpdatedAt:   nil,
			DeletedAt:   nil,
		},
		Description: req.Description,
		AuthorName:  req.AuthorName,
		AuthorLink:  req.AuthorLink,
		Homepage:    req.Homepage,
		ReleaseDate: req.ReleaseDate,
		CoverImage:  req.CoverImage,
	}

	err := h.service.CreateBook(ctx, book)
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonCreated(book)
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
