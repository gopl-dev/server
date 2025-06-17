package endpoint

import h "github.com/gopl-dev/server/server/handler"

func (r *Router) ProtectedWebEndpoints() {
	r.GET("/users/logout/", h.UserLogout)
}
