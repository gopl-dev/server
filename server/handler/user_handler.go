package handler

import (
	"net/http"

	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
	"github.com/gopl-dev/server/server/request"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req request.RegisterUser
	h := handleJSON(w, r, &req)
	if h.Aborted() {
		return
	}

	user, err := service.RegisterUser(r.Context(), req.ToParams())
	if err != nil {
		h.Abort(err)
		return
	}

	h.jsonOK(user)
}

func RegisterUserViewForm(w http.ResponseWriter, r *http.Request) {
	head := page.HeadData{
		Title: "Register",
	}

	page := page.Page{
		Name: "Register",
		Head: page.Head{},
		Body:
	}

	data := layout.Data{
		Title: "Register",
		Head:  page.Head(head),
		Body:  page.RegisterUserForm(),
		User:  nil, // TODO! resolve user
	}

	render(r.Context(), w, layout.Default(data))
}
