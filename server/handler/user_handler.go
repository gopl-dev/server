package handler

import (
	"log"
	"net/http"

	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
)

// UserSignUp is the API handler for user registration.
//
//	@ID			UserSignUp
//	@Summary	User registration
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.UserSignUp	true	"Request body"
//	@Success	200		{object}	response.Status
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/users/sign-up/ [post]
//	@Security	ApiKeyAuth
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

// UserSignIn is the API handler for the user login endpoint.
//
//	@ID			UserSignIn
//	@Summary	User auth
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.UserSignIn	true	"Request body"
//	@Success	200		{object}	response.UserSignIn
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/users/sign-in/ [post]
//	@Security	ApiKeyAuth
//
// TODO either email or username can be used to login.
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

// ConfirmEmail is the API handler for confirming a user's email address via a confirmation code.
//
//	@ID			ConfirmEmail
//	@Summary	Confirm email
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.ConfirmEmail	true	"Request body"
//	@Success	200		{object}	response.Status
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/users/confirm-email/ [post]
//	@Security	ApiKeyAuth
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

// UserSignUpView renders the static HTML form for user registration.
func (h *Handler) UserSignUpView(w http.ResponseWriter, r *http.Request) {
	renderTempl(r.Context(), w, layout.Default(layout.Data{
		Title: "Sign up",
		Body:  page.UserSignUpForm(), // Assumes page.UserSignUpForm is the templ component for the form
		User:  nil,                   // TODO! resolve user (Placeholder for authenticated user object, if required)
	}))
}

// UserSignOut handles user log-out by clearing the session cookie and deleting the session
// record from the database.
//
//	@ID			UserSignOut
//	@Summary	Logout
//	@Tags		users
//	@Produce	json
//	@Success	200		{object}	response.Status
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/users/confirm-email/ [post]
//	@Security	ApiKeyAuth
func (h *Handler) UserSignOut(w http.ResponseWriter, r *http.Request) {
	// Removes the session cookie from the client.
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

// ConfirmEmailView renders the static HTML form page where a user can manually enter
// an email confirmation code.
func (h *Handler) ConfirmEmailView(w http.ResponseWriter, r *http.Request) {
	renderTempl(r.Context(), w, layout.Default(layout.Data{
		Title: "Confirm email",
		Body:  page.ConfirmEmailForm(), // Assumes page.ConfirmEmailForm is the templ component
		User:  nil,                     // TODO! resolve user
	}))
}

// UserSignInView renders the static HTML form for user login.
// It is a wrapper around the RenderUserSignInPage helper.
func (h *Handler) UserSignInView(w http.ResponseWriter, r *http.Request) {
	RenderUserSignInPage(w, r, "/")
}
