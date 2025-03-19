package server

import h "github.com/gopl-dev/server/app/server/handler"

func (r *Router) RegisterPublicWebRoutes() {
	r.GET("/", h.Home)
}
