package endpoint

// PublicWebEndpoints registers all public-facing web endpoints that do not require authentication.
func (r *Router) PublicWebEndpoints() {
	r.GET("/", r.handler.Home)

	// User authentication and registration
	r.GET("/users/sign-up/", r.handler.UserSignUpView)
	r.GET("/users/confirm-email/", r.handler.ConfirmEmailView)
	r.GET("/users/sign-in/", r.handler.UserSignInView)

	r.GET("/password-reset/", r.handler.PasswordResetRequestView)
	r.GET("/password-reset/{token}/", r.handler.PasswordResetConfirmView)
}
