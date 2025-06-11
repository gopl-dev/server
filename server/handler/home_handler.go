package handler

import (
	"net/http"

	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/frontend"
	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
)

func Home(w http.ResponseWriter, r *http.Request) {
	renderTempl(r.Context(), w, layout.Default(layout.Data{
		Title: "Welcome!",
		Body:  page.Home(),
		User:  frontend.NewUser(service.UserFromContext(r.Context())),
	}))
}
