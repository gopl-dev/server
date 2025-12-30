package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
	"github.com/markbates/goth/gothic"
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
	ctx, span := h.tracer.Start(r.Context(), "UserSignUp")
	defer span.End()

	var req request.UserSignUp

	res := handleJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	_, err := h.service.RegisterUser(ctx, req.Username, req.Email, req.Password)
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
	ctx, span := h.tracer.Start(r.Context(), "UserSignIn")
	defer span.End()

	var req request.UserSignIn

	res := handleJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	user, token, err := h.service.AuthenticateUser(ctx, req.Email, req.Password)
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
	ctx, span := h.tracer.Start(r.Context(), "ConfirmEmail")
	defer span.End()

	var req request.ConfirmEmail
	res := handleJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	err := h.service.ConfirmEmail(ctx, req.Code)
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonSuccess()
}

// UserSignUpView renders the static HTML form for user registration.
func (h *Handler) UserSignUpView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "UserSignUpView")
	defer span.End()

	if ds.UserFromContext(ctx) != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	renderDefaultLayout(ctx, w, layout.Data{
		Title: "Sign up",
		Body:  page.UserSignUpForm(),
	})
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
//	@Router		/users/sign-out/ [post]
//	@Security	ApiKeyAuth
func (h *Handler) UserSignOut(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "UserSignOut")
	defer span.End()

	clearSessionCookie(w)

	session := ds.UserSessionFromContext(ctx)
	if session != nil {
		err := h.service.DeleteUserSession(ctx, session.ID)
		if err != nil {
			log.Println("delete user session: " + err.Error())
		}
	}

	if request.IsJSON(r) {
		jsonOK(w, response.Success)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// ChangePassword handles the API request for an authenticated user to change their password.
//
//	@ID			ChangePassword
//	@Summary	Change user password
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.ChangePassword	true	"Old and new passwords"
//	@Success	200		{object}	response.Status
//	@Failure	401		{object}	Error "Unauthorized"
//	@Failure	422		{object}	Error "Validation error or incorrect old password"
//	@Failure	500		{object}	Error
//	@Router		/users/password/ [post]
//	@Security	ApiKeyAuth
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ChangePassword")
	defer span.End()

	var req request.ChangePassword
	user, res := handleAuthorizedJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	err := h.service.ChangeUserPassword(ctx, user.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonSuccess()
}

// ConfirmEmailView renders the static HTML form page where a user can enter
// an email confirmation code.
func (h *Handler) ConfirmEmailView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConfirmEmailView")
	defer span.End()

	renderDefaultLayout(ctx, w, layout.Data{
		Title: "Confirm email",
		Body:  page.ConfirmEmailForm(),
	})
}

// UserSettingsView renders the static HTML page where a user can manually enter
// an email confirmation code.
func (h *Handler) UserSettingsView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "UserSettingsView")
	defer span.End()

	renderDefaultLayout(ctx, w, layout.Data{
		Title: "Settings",
		Body:  page.UserSettings(),
	})
}

// ChangePasswordView renders the static HTML page where a user can manually enter
// an email confirmation code.
func (h *Handler) ChangePasswordView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ChangePasswordView")
	defer span.End()

	renderDefaultLayout(ctx, w, layout.Data{
		Title: "Change password",
		Body:  page.ChangePasswordForm(),
	})
}

// RequestEmailChangeView renders the page with the form to request an email change.
func (h *Handler) RequestEmailChangeView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RequestEmailChangeView")
	defer span.End()

	renderDefaultLayout(ctx, w, layout.Data{
		Title: "Change Email",
		Body:  page.ChangeEmailForm(),
	})
}

// RequestEmailChange handles the request for an email change.
//
//	@ID			EmailChangeRequest
//	@Summary	Request to change user email
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.EmailChangeRequest	true	"Old and new passwords"
//	@Success	200		{object}	response.Status
//	@Failure	401		{object}	Error "Unauthorized"
//	@Failure	422		{object}	Error "Validation error"
//	@Failure	500		{object}	Error
//	@Router		/users/email/ [post]
//	@Security	ApiKeyAuth
func (h *Handler) RequestEmailChange(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "EmailChangeRequest")
	defer span.End()

	var req request.EmailChangeRequest
	user, res := handleAuthorizedJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	err := h.service.CreateChangeEmailRequest(ctx, user.ID, req.Email)
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonSuccess()
}

// ConfirmEmailChange handles the confirmation for an email change.
//
//	@ID			EmailChangeConfirm
//	@Summary	Confirm changing user email
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.EmailChangeRequest	true	"Old and new passwords"
//	@Success	200		{object}	response.Status
//	@Failure	401		{object}	Error "Unauthorized"
//	@Failure	422		{object}	Error "Validation error"
//	@Failure	500		{object}	Error
//	@Router		/users/email/ [post]
//	@Security	ApiKeyAuth
func (h *Handler) ConfirmEmailChange(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "EmailChangeConfirm")
	defer span.End()

	var req request.EmailChangeConfirm
	res := handleJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	err := h.service.ConfirmEmailChange(ctx, req.Token)
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonSuccess()
}

// ConfirmEmailChangeView handles the confirmation link for an email change.
func (h *Handler) ConfirmEmailChangeView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ConfirmEmailChangeView")
	defer span.End()

	token := r.PathValue("token")
	err := h.service.ConfirmEmailChange(ctx, token)
	if errors.Is(err, service.ErrInvalidChangeEmailToken) {
		renderDefaultLayout(ctx, w, layout.Data{
			Title: "Change email",
			Body:  page.Err422(err.Error()),
		})
		return
	}
	if err != nil {
		renderDefaultLayout(ctx, w, layout.Data{
			Title: "Change email",
			Body:  page.Err500(err.Error()),
		})
		return
	}

	renderDefaultLayout(ctx, w, layout.Data{
		Title: "Change email",
		Body:  page.SuccessMessage("Your email successfully changed!"),
	})
}

// UserSignInView renders the static HTML form for user login.
// It is a wrapper around the RenderUserSignInPage helper.
func (h *Handler) UserSignInView(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "UserSignInView")
	defer span.End()

	RenderUserSignInPage(w, r, "/")
}

// ChangeUsernameView renders the page with the form to change username.
func (h *Handler) ChangeUsernameView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ChangeUsernameView")
	defer span.End()

	renderDefaultLayout(ctx, w, layout.Data{
		Title: "Change Username",
		Body:  page.ChangeUsernameForm(),
	})
}

// ChangeUsername handles the API request for an authenticated user to change their username.
//
//	@ID			ChangeUsername
//	@Summary	Change username
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.ChangeUsername	true	"New username and password"
//	@Success	200		{object}	response.Status
//	@Failure	401		{object}	Error "Unauthorized"
//	@Failure	422		{object}	Error "Validation error or incorrect password"
//	@Failure	500		{object}	Error
//	@Router		/users/username/ [post]
//	@Security	ApiKeyAuth
func (h *Handler) ChangeUsername(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ChangeUsername")
	defer span.End()

	var req request.ChangeUsername
	user, res := handleAuthorizedJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	err := h.service.ChangeUsername(ctx, service.ChangeUsernameInput{
		UserID:      user.ID,
		NewUsername: req.Username,
		Password:    req.Password,
	})
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonSuccess()
}

// DeleteUserView renders the page with the form to delete user account.
func (h *Handler) DeleteUserView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "DeleteUserView")
	defer span.End()

	renderDefaultLayout(ctx, w, layout.Data{
		Title: "Delete Account",
		Body:  page.DeleteUserForm(),
	})
}

// OAuthStart ...
func (h *Handler) OAuthStart(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "OAuthStart")
	defer span.End()

	oauthUser, err := gothic.CompleteUserAuth(w, r)
	if err == nil {
		token, err := h.service.AuthenticateOAuthUser(ctx, oauthUser)
		if err != nil {
			Abort(w, err)
			return
		}

		setSessionCookie(w, token)

		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	gothic.BeginAuthHandler(w, r)
}

// OAuthComplete ...
func (h *Handler) OAuthComplete(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "OAuthComplete")
	defer span.End()

	oauthUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		Abort(w, err)
		return
	}

	token, err := h.service.AuthenticateOAuthUser(ctx, oauthUser)
	if err != nil {
		Abort(w, err)
		return
	}

	setSessionCookie(w, token)

	http.Redirect(w, r, "/", http.StatusFound)
}

// DeleteUser handles the API request for an authenticated user to delete their account.
//
//	@ID			DeleteUser
//	@Summary	Delete user account
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.DeleteUser	true	"Password"
//	@Success	200		{object}	response.Status
//	@Failure	401		{object}	Error "Unauthorized"
//	@Failure	422		{object}	Error "Validation error or incorrect password"
//	@Failure	500		{object}	Error
//	@Router		/users/ [delete]
//	@Security	ApiKeyAuth
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "DeleteUser")
	defer span.End()

	var req request.DeleteUser
	user, res := handleAuthorizedJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	err := h.service.DeleteUser(ctx, user.ID, req.Password)
	if err != nil {
		res.Abort(err)
		return
	}

	clearSessionCookie(w)
	res.jsonSuccess()
}
