package handler

import (
	"net/http"

	"github.com/gopl-dev/server/app/web"
	"github.com/gopl-dev/server/app/web/layout"
	"github.com/gopl-dev/server/app/web/page"
)

func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	head := web.HeadData{
		Title: "Hey, Bro!",
	}
	data := layout.Data{
		Head: web.Head(head),
		Body: page.Home(),
	}

	err := layout.Default(data).Render(r.Context(), w)
	if err != nil {
		abort(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
