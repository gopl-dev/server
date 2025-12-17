package endpoint

// ProtectedWebEndpoints ...
func (r *Router) ProtectedWebEndpoints() {
	r.GET("/users/sign-out/", r.handler.UserSignOut)
	r.GET("/users/settings/", r.handler.UserSettingsView)
	r.GET("/change-password/", r.handler.ChangePasswordView)
	r.GET("/change-email/", r.handler.RequestEmailChangeView)
	r.GET("/change-email/{token}/", r.handler.ConfirmEmailChangeView)
}
