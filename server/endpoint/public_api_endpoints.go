package endpoint

import (
	h "github.com/gopl-dev/server/server/handler"
)

func (r *Router) PublicApiEndpoints() {
	r.GET("status", h.StatusHandler)

	r.Group("users").
		POST("/register/", h.RegisterUser).
		POST("/login/", h.LoginUser).
		POST("/confirm-email/", h.ConfirmEmail)

	r.Group("books").
		GET("/", h.FilterBooks).
		GET("{book_id}", h.GetBookByID)

}
