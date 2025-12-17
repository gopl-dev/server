package handler

import (
	"net/http"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/frontend"
	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
)

// Home sweet home.
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "Home")
	defer span.End()

	renderTempl(ctx, w, layout.Default(layout.Data{
		Title: "Welcome!",
		Body:  page.Home(),
		User:  frontend.NewUser(ds.UserFromContext(r.Context())),
	}))
}
