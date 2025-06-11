package middleware

import (
	"net/http"

	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/server/endpoint"
	"github.com/gopl-dev/server/server/handler"
)

func ResolveUserFromCookie(next endpoint.Handler) endpoint.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		jwt := handler.GetSessionFromCookie(r)
		if jwt != "" {
			user, session, err := service.GetUserAndSessionFromJWT(r.Context(), jwt)
			if err != nil {
				next(w, r)
				return
			}

			ctx := service.UserToContext(r.Context(), user)
			ctx = service.UserSessionToContext(ctx, session)

			r = r.WithContext(ctx)
		}

		next(w, r)
	}
}
