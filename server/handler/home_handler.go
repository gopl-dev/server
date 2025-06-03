package handler

import (
	"net/http"

	"github.com/gopl-dev/server/frontend"
	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
)

func Home(w http.ResponseWriter, r *http.Request) {
	head := frontend.HeadData{
		Title: "Welcome!",
	}
	data := layout.Data{
		Title: "Home",
		Head:  frontend.Head(head),
		Body:  page.Home(),
		User:  nil, // TODO resolve user
	}

	render(r.Context(), w, layout.Default(data))
}
