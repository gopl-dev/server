package endpoint

// ProtectedAPIEndpoints is endpoints that require user auth.
func (r *Router) ProtectedAPIEndpoints() {
	r.POST("/users/change-password/", r.handler.ChangePassword)
}
