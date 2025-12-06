package endpoint

func (r *Router) PublicApiEndpoints() {
	r.GET("status", r.handler.StatusHandler)

	r.Group("users").
		POST("/register/", r.handler.RegisterUser).
		POST("/login/", r.handler.LoginUser).
		POST("/confirm-email/", r.handler.ConfirmEmail)

	r.Group("books").
		GET("/", r.handler.FilterBooks).
		GET("{book_id}", r.handler.GetBookByID)

}
