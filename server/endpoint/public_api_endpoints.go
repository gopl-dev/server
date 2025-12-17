package endpoint

// PublicAPIEndpoints ...
func (r *Router) PublicAPIEndpoints() {
	r.GET("status/", r.handler.ServerStatus)

	r.Group("users").
		POST("sign-up/", r.handler.UserSignUp).
		POST("sign-in/", r.handler.UserSignIn).
		POST("confirm-email/", r.handler.ConfirmEmail).
		POST("password-reset-request/", r.handler.PasswordResetRequest).
		POST("password-reset/", r.handler.PasswordResetConfirm)

	r.Group("books").
		GET("/", r.handler.FilterBooks).
		GET("{book_id}/", r.handler.GetBookByID)
}
