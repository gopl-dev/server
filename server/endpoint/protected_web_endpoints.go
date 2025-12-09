package endpoint

// ProtectedWebEndpoints ...
func (r *Router) ProtectedWebEndpoints() {
	r.GET("/users/logout/", r.handler.UserSignOut)
}
