package endpoint

import h "github.com/gopl-dev/server/server/handler"

func (r *Router) PublicWebEndpoints() {
	r.GET("/", h.Home)
	r.GET("/users/register/", h.RegisterUserView)
	r.GET("/users/confirm-email/", h.ConfirmEmailView)
	r.GET("/users/login/", h.LoginUserView)
}
