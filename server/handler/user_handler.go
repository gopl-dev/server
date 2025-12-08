package handler

import (
	"log"
	"net/http"

	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
)

func (h *Handler) UserSignUp(w http.ResponseWriter, r *http.Request) {
	var req request.UserSignUp
	res := handleJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	_, err := h.service.RegisterUser(r.Context(), req.ToParams())
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonSuccess()
}

// UserSignIn is a handler for the user login endpoint.
// TODO either email or username can be used to login
func (h *Handler) UserSignIn(w http.ResponseWriter, r *http.Request) {
	var req request.UserSignIn
	res := handleJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	user, token, err := h.service.LoginUser(r.Context(), req.Email, req.Password)
	if err != nil {
		res.Abort(err)
		return
	}

	setSessionCookie(w, token)

	res.jsonOK(response.UserSignIn{
		ID:       user.ID,
		Username: user.Username,
		Token:    token,
	})
}

func (h *Handler) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	var req request.ConfirmEmail
	res := handleJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	err := h.service.ConfirmEmail(r.Context(), req.Code)
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonSuccess()
}

func (h *Handler) UserSignUpView(w http.ResponseWriter, r *http.Request) {
	renderTempl(r.Context(), w, layout.Default(layout.Data{
		Title: "Sign up",
		Body:  page.UserSignUpForm(),
		User:  nil, // TODO! resolve user
	}))
}

func (h *Handler) UserSignOut(w http.ResponseWriter, r *http.Request) {
	clearSessionCookie(w)

	ctx := r.Context()
	session := h.service.UserSessionFromContext(ctx)
	if session != nil {
		err := h.service.DeleteUserSession(ctx, session.ID)
		if err != nil {
			log.Println("delete user session: " + err.Error())
		}
	}

	if isJSON(r) {
		jsonOK(w, response.Success)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) ConfirmEmailView(w http.ResponseWriter, r *http.Request) {
	renderTempl(r.Context(), w, layout.Default(layout.Data{
		Title: "Confirm email",
		Body:  page.ConfirmEmailForm(),
		User:  nil, // TODO! resolve user
	}))
}

func (h *Handler) UserSignInView(w http.ResponseWriter, r *http.Request) {
	RenderUserSignInPage(w, r, "/")
}
