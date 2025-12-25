package endpoint

// ProtectedAPIEndpoints is endpoints that require user auth.
func (r *Router) ProtectedAPIEndpoints() {
	r.PUT("/users/password/", r.handler.ChangePassword)
	r.POST("/users/email/", r.handler.RequestEmailChange)
	r.PUT("/users/email/", r.handler.ConfirmEmailChange)
	r.PUT("/users/username/", r.handler.ChangeUsername)
	r.DELETE("/users/", r.handler.DeleteUser)
}
