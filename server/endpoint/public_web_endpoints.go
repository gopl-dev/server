package endpoint

func (r *Router) PublicWebEndpoints() {
	r.GET("/", r.handler.Home)
	r.GET("/users/register/", r.handler.RegisterUserView)
	r.GET("/users/confirm-email/", r.handler.ConfirmEmailView)
	r.GET("/users/login/", r.handler.LoginUserView)
}
