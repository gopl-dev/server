package handler

import (
	"net/http"

	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req request.RegisterUser
	h := handleJSON(w, r, &req)
	if h.Aborted() {
		return
	}

	_, err := service.RegisterUser(r.Context(), req.ToParams())
	if err != nil {
		h.Abort(err)
		return
	}

	h.jsonSuccess()
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var req request.UserLogin
	h := handleJSON(w, r, &req)
	if h.Aborted() {
		return
	}

	user, token, err := service.LoginUser(r.Context(), req.Email, req.Password)
	if err != nil {
		h.Abort(err)
		return
	}

	setSessionCookie(w, token)

	h.jsonOK(response.LoginUser{
		ID:       user.ID,
		Username: user.Username,
		Token:    token,
	})
}

func ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	var req request.ConfirmEmail
	h := handleJSON(w, r, &req)
	if h.Aborted() {
		return
	}

	err := service.ConfirmEmail(r.Context(), req.Code)
	if err != nil {
		h.Abort(err)
		return
	}

	h.jsonSuccess()
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

func LoginUserView(w http.ResponseWriter, r *http.Request) {
	renderTempl(r.Context(), w, layout.Default(layout.Data{
		Title: "Login",
		Body:  page.UserLoginForm(),
		User:  nil, // TODO! resolve user
	}))
}
