package middleware

import (
	"net/http"

	"github.com/gopl-dev/server/server/handler"
)

// RequestPage resolves a page from the request path and injects it into the request context.
func (mw *Middleware) RequestPage(next handler.Fn) handler.Fn {
	return func(w http.ResponseWriter, r *http.Request) {
		pageID := r.PathValue("id")

		page, err := mw.service.GetPageByPublicID(r.Context(), pageID)
		if err != nil {
			handler.Abort(w, r, err)
			return
		}

		ctx := page.ToContext(r.Context())
		r = r.WithContext(ctx)

		next(w, r)
	}
}
