package middleware

import (
	"net/http"

	"github.com/gopl-dev/server/server/handler"
)

// RequestBook resolves a book from the request path and injects it into the request context.
//
// The middleware reads the "id" path parameter, which may contain either an internal
// UUID-based identifier or a public book identifier.
// On successful resolution, the book is stored in the request context and passed to
// the next handler.
func (mw *Middleware) RequestBook(next handler.Fn) handler.Fn {
	return func(w http.ResponseWriter, r *http.Request) {
		bookID := r.PathValue("id")

		book, err := mw.service.GetBookByRef(r.Context(), bookID)
		if err != nil {
			handler.Abort(w, r, err)
			return
		}

		ctx := book.ToContext(r.Context())
		r = r.WithContext(ctx)

		next(w, r)
	}
}
