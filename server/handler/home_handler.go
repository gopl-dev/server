package handler

import (
	"net/http"

	"github.com/gopl-dev/server/web"
	"github.com/gopl-dev/server/web/layout"
	"github.com/gopl-dev/server/web/page"
)

func Home(w http.ResponseWriter, r *http.Request) {
	head := web.HeadData{
		Title: "Welcome!",
	}
	data := layout.Data{
		Title: "Home",
		Head:  web.Head(head),
		Body:  page.Home(),
		User:  nil, // TODO resolve user
	}

	render(r.Context(), w, layout.Default(data))
}
