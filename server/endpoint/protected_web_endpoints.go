package endpoint

// ProtectedWebEndpoints registers web endpoints that require user authentication.
func (r *Router) ProtectedWebEndpoints() {
	// users
	r.GET("/users/sign-out/", r.handler.UserSignOut)
	r.GET("/users/settings/", r.handler.UserSettingsView)
	r.GET("/change-password/", r.handler.ChangePasswordView)
	r.GET("/change-email/", r.handler.RequestEmailChangeView)
	r.GET("/change-email/{token}/", r.handler.ConfirmEmailChangeView)
	r.GET("/change-username/", r.handler.ChangeUsernameView)
	r.GET("/delete-account/", r.handler.DeleteUserView)

	// books
	r.GET("/add-book/", r.handler.CreateBookView)
	r.Group("/edit-book/{id}/", r.mw.RequestBook).
		GET("/", r.handler.EditBookView)

	// pages
	r.GET("/add-page/", r.handler.CreatePageView)
	r.Group("/edit-page/{id}/", r.mw.RequestPage).
		GET("/", r.handler.EditPageView)
}
