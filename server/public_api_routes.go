package server

import (
	h "github.com/gopl-dev/server/server/handler"
)

func registerPublicApiRoutes(r *Router) {
	r.GET("status", h.StatusHandler)

	r.Group("users").
		POST("/register/", h.RegisterUser).
		POST("/confirm-email/", h.ConfirmEmail)

	r.Group("books").
		GET("/", h.FilterBooks).
		GET("{book_id}", h.GetBookByID)

}
