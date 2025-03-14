package api

import (
	"github.com/gopl-dev/server/app/api/handler"
)

func registerPublicApiRoutes(r *Router) {
	r.GET("status", handler.StatusHandler)

	r.Group("books").
		GET("/", handler.FilterBooks).
		GET("{book_id}", handler.GetBookByID)

}
