package api

import (
	h "github.com/gopl-dev/server/app/api/handler"
)

func registerPublicApiRoutes(r *Router) {
	r.GET("status", h.StatusHandler)

	r.Group("users").
		POST("/register/", h.RegisterUser)

	r.Group("books").
		GET("/", h.FilterBooks).
		GET("{book_id}", h.GetBookByID)

}
