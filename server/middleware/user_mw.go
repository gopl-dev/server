package middleware

import (
	"net/http"
	"slices"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/server/handler"
)

// ResolveUserFromCookie is a middleware that attempts to find a user's session token
// in an HTTP cookie, validate the token and the session, and if successful,
// adds the authenticated user and session objects to the request's context.
func (mw *Middleware) ResolveUserFromCookie(next handler.Fn) handler.Fn {
	return func(w http.ResponseWriter, r *http.Request) {
		token := handler.GetSessionFromCookie(r)
		if token != "" {
			user, session, err := mw.service.GetUserAndSessionFromJWT(r.Context(), token)
			if err != nil {
				next(w, r)
				return
			}

			user.IsAdmin = slices.Contains(app.Config().Admins, user.ID.String())

			ctx := user.ToContext(r.Context())
			ctx = session.ToContext(ctx)

			err = mw.service.ProlongUserSession(ctx, session.ID)
			if err != nil {
				handler.Abort(w, r, err)
				return
			}

			r = r.WithContext(ctx)
		}

		next(w, r)
	}
}

// UserAuth is a middleware that enforces user authentication.
// For API (JSON) requests, it returns a JSON 401 Unauthorized error if the user is not authenticated.
// For all other requests (e.g., web pages), it renders login form.
func (mw *Middleware) UserAuth(next handler.Fn) handler.Fn {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ds.UserFromContext(r.Context())
		if user == nil {
			if handler.ShouldServeJSON(r) {
				handler.Abort(w, r, app.ErrUnauthorized())
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

// AdminOnly is a middleware that restricts access to admin users only.
func (mw *Middleware) AdminOnly(next handler.Fn) handler.Fn {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ds.UserFromContext(r.Context())
		if user == nil || !user.IsAdmin {
			if handler.ShouldServeJSON(r) {
				handler.Abort(w, r, app.ErrUnauthorized())
				return
			}
		}

		next(w, r)
	}
}
