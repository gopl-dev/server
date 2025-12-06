package endpoint

func (r *Router) ProtectedWebEndpoints() {
	r.GET("/users/logout/", r.handler.UserLogout)
}
