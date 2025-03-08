package server

import (
	"github.com/gopl-dev/server/app/server/handler"
)

func registerPublicApiRoutes(r *Router) {
	r.GET("status", handler.StatusHandler)

	r.Group("books").
		GET("/", handler.FilterBooks).
		GET("{id}", handler.GetBookByID)

	r.GET("content/import", handler.ImportContentFromGitHubRepo)
}
