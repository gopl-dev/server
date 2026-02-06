package handler

import (
	"net/http"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/frontend"
	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
)

// Dashboard salty dashboard.
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "Dashboard")
	defer span.End()

	user := ds.UserFromContext(ctx)
	if user == nil || !user.IsAdmin {
		Abort(w, r, app.ErrUnauthorized())
		return
	}

	renderTempl(ctx, w, layout.Dashboard(layout.Data{
		Title: "Dashboard",
		Body:  page.Home(),
		User:  frontend.NewUser(ds.UserFromContext(r.Context())),
	}))
}
