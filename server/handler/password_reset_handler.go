package handler

import (
	"net/http"

	"github.com/gopl-dev/server/frontend/layout"
	"github.com/gopl-dev/server/server/request"
)

// PasswordResetRequestView renders the page with the form to request a password reset.
func (h *Handler) PasswordResetRequestView(w http.ResponseWriter, r *http.Request) {
	renderTempl(r.Context(), w, layout.Default(layout.Data{
		Title: "Request Password Reset",
		//Body:  page.PasswordRequestResetForm(),
	}))
}

// PasswordResetRequest handles the form submission for requesting a password reset.
// It calls the service to generate and send a reset token.
func (h *Handler) PasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	var req request.PasswordRequestReset
	res := handleJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	err := h.service.RequestPasswordReset(r.Context(), req.Email)
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonSuccess()
}

// PasswordResetConfirmView renders the page with the form to reset the password, using a token from the query string.
func (h *Handler) PasswordResetConfirmView(w http.ResponseWriter, r *http.Request) {
	// token := r.URL.Query().Get("token")
	renderTempl(r.Context(), w, layout.Default(layout.Data{
		Title: "Reset Your Password",
		// Body:  page.PasswordResetForm(token),
	}))
}

// PasswordResetConfirm handles the form submission for resetting the password.
// It validates the token and new password, and calls the service to perform the reset.
func (h *Handler) PasswordResetConfirm(w http.ResponseWriter, r *http.Request) {
	var req request.PasswordReset
	res := handleJSON(w, r, &req)
	if res.Aborted() {
		return
	}

	err := h.service.ResetPassword(r.Context(), req.Token, req.Password)
	if err != nil {
		res.Abort(err)
		return
	}

	res.jsonSuccess()
}
