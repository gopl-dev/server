package server

import h "github.com/gopl-dev/server/server/handler"

func (r *Router) RegisterPublicWebRoutes() {
	r.GET("/", h.Home)
	r.GET("/users/register/", h.RegisterUserViewForm)
}
