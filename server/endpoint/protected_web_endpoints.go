package endpoint

// ProtectedWebEndpoints ...
func (r *Router) ProtectedWebEndpoints() {
	r.GET("/users/logout/", r.handler.UserSignOut)
	r.GET("/users/settings/", r.handler.UserSettingsView)
	r.GET("/change-password/", r.handler.ChangePasswordView)
}
