package middleware

import (
	"net/http"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/server/endpoint"
	"github.com/gopl-dev/server/server/handler"
	"github.com/gopl-dev/server/server/request"
)

// ResolveUserFromCookie is a middleware that attempts to find a user's session token
// in an HTTP cookie, validate the token and the session, and if successful,
// adds the authenticated user and session objects to the request's context.
func (mw *Middleware) ResolveUserFromCookie(next endpoint.Handler) endpoint.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		jwt := handler.GetSessionFromCookie(r)
		if jwt != "" {
			user, session, err := mw.service.GetUserAndSessionFromJWT(r.Context(), jwt)
			if err != nil {
				next(w, r)
				return
			}

			ctx := user.ToContext(r.Context())
			ctx = session.ToContext(ctx)

			r = r.WithContext(ctx)
		}

		next(w, r)
	}
}

// UserAuth is a middleware that enforces user authentication.
// For API (JSON) requests, it returns a JSON 401 Unauthorized error if the user is not authenticated.
// For all other requests (e.g., web pages), it renders login form.
func (mw *Middleware) UserAuth(next endpoint.Handler) endpoint.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ds.UserFromContext(r.Context())
		if user == nil {
			if request.IsJSON(r) {
				handler.Abort(w, app.ErrUnauthorized())
				return
			}

			redirectTo := r.URL.Path
			if redirectTo == "/users/logout/" {
				redirectTo = "/"
			}

			handler.RenderUserSignInPage(w, r, redirectTo)
			return
		}

		next(w, r)
	}
}
