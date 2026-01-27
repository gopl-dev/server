package handler

import (
	"errors"
	"net/http"

	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/frontend/page"
	"github.com/gopl-dev/server/server/request"
)

// PasswordResetRequestView renders the page with the form to request a password reset.
func (h *Handler) PasswordResetRequestView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PasswordResetRequestView")
	defer span.End()

	renderTempl(ctx, w, layout.Default(layout.Data{
		Title: "Request Password Reset",
		Body:  page.PasswordResetRequestForm(),
	}))
}

// PasswordResetRequest handles the form submission for requesting a password reset.
//
//	@ID			PasswordResetRequest
//	@Summary	Initiate password reset
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PasswordResetRequest	true	"Request body"
//	@Success	200		{object}	response.Status
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/users/password-reset-request/ [post]
//	@Security	ApiKeyAuth
func (h *Handler) PasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PasswordResetRequest")
	defer span.End()

	var req request.PasswordResetRequest
	res := handleJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	err := h.service.CreatePasswordResetRequest(ctx, req.Email)
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonSuccess()
}

// PasswordResetConfirmView renders the page with the form to reset the password.
func (h *Handler) PasswordResetConfirmView(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PasswordResetConfirmView")
	defer span.End()

	token := r.PathValue("token")

	_, err := h.service.FindPasswordResetByToken(ctx, token)
	if errors.Is(err, service.ErrInvalidPasswordResetToken) {
		RenderDefaultLayout(ctx, w, layout.Data{
			Title: "Reset Your Password",
			Body:  page.Err422(err.Error()),
		})
		return
	}
	if err != nil {
		RenderDefaultLayout(ctx, w, layout.Data{
			Title: "Reset Your Password",
			Body:  page.Err500(err.Error()),
		})
		return
	}

	RenderDefaultLayout(ctx, w, layout.Data{
		Title: "Reset Your Password",
		Body:  page.PasswordResetForm(token),
	})
}

// PasswordResetConfirm handles the form submission for resetting the password.
//
//	@ID			PasswordResetConfirm
//	@Summary	Password reset
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PasswordReset	true	"Request body"
//	@Success	200		{object}	response.Status
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/users/password-reset/ [post]
//	@Security	ApiKeyAuth
func (h *Handler) PasswordResetConfirm(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PasswordResetConfirm")
	defer span.End()

	var req request.PasswordReset
	res := handleJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	err := h.service.ResetPassword(ctx, req.Token, req.Password)
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonSuccess()
}
