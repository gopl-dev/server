package middleware

import (
	"net/http"

	"github.com/gopl-dev/server/server/endpoint"
	"github.com/gopl-dev/server/server/handler"
)

func (mw *Middleware) ResolveUserFromCookie(next endpoint.Handler) endpoint.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		jwt := handler.GetSessionFromCookie(r)
		if jwt != "" {
			user, session, err := mw.service.GetUserAndSessionFromJWT(r.Context(), jwt)
			if err != nil {
				next(w, r)
				return
			}

			ctx := mw.service.UserToContext(r.Context(), user)
			ctx = mw.service.UserSessionToContext(ctx, session)

			r = r.WithContext(ctx)
		}

		next(w, r)
	}
}

func (mw *Middleware) UserAuthWeb(next endpoint.Handler) endpoint.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		user := mw.service.UserFromContext(r.Context())
		if user == nil {
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
