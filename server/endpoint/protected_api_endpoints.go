package endpoint

// ProtectedAPIEndpoints is endpoints that require user auth.
func (r *Router) ProtectedAPIEndpoints() {
	// users
	r.PUT("/users/password/", r.handler.ChangePassword)
	r.POST("/users/email/", r.handler.RequestEmailChange)
	r.PUT("/users/email/", r.handler.ConfirmEmailChange)
	r.PUT("/users/username/", r.handler.ChangeUsername)
	r.DELETE("/users/", r.handler.DeleteUser)

	// books
	r.POST("/books/", r.handler.CreateBook)
	r.Group("/books/{id}/", r.mw.RequestBook).
		GET("/edit/", r.handler.GetBookEditState)

	// files
	r.POST("/files/", r.handler.UploadFile)
	r.DELETE("/files/{id}/", r.handler.DeleteFile)
}
