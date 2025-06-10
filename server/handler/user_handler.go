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

func ConfirmEmail(w http.ResponseWriter, r *http.Request) {
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

func RegisterUserView(w http.ResponseWriter, r *http.Request) {
	renderTempl(r.Context(), w, layout.Default(layout.Data{
		Title: "Register",
		Body:  page.RegisterUserForm(),
		User:  nil, // TODO! resolve user
	}))
}

func ConfirmEmailView(w http.ResponseWriter, r *http.Request) {
	renderTempl(r.Context(), w, layout.Default(layout.Data{
		Title: "Confirm email",
		Body:  page.ConfirmEmailForm(),
		User:  nil, // TODO! resolve user
	}))
}
