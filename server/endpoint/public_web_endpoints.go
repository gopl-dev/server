package endpoint

func (r *Router) PublicWebEndpoints() {
	r.GET("/", r.handler.Home)
	r.GET("/users/sign-up/", r.handler.UserSignUpView)
	r.GET("/users/confirm-email/", r.handler.ConfirmEmailView)
	r.GET("/users/sign-in/", r.handler.UserSignInView)
}
