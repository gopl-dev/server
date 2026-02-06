package endpoint

// ProtectedAPIEndpoints registers API routes that require authentication.
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
		PUT("/", r.handler.UpdateBook).
		GET("/edit/", r.handler.GetBookEditState).
		PUT("/approve/", r.handler.ApproveNewBook).
		PUT("/reject/", r.handler.RejectNewBook)

	// pages
	r.POST("/pages/", r.handler.CreatePage)
	r.Group("/pages/{id}/", r.mw.RequestPage).
		PUT("/", r.handler.UpdatePage).
		GET("/edit/", r.handler.GetPageEditState)

	// files
	r.POST("/files/", r.handler.UploadFile)
	r.DELETE("/files/{id}/", r.handler.DeleteFile)

	// change requests
	r.Group("/change-requests/", r.mw.AdminOnly).
		GET("/", r.handler.FilterChangeRequests).
		PUT("/{id}/apply/", r.handler.ApplyChangeRequest).
		PUT("/{id}/reject/", r.handler.RejectChangeRequest)
}
