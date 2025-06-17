package server

import (
	"net"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/server/endpoint"
	"github.com/gopl-dev/server/server/middleware"
)

func NewServer() *http.Server {
	conf := app.Config().Server

	r := endpoint.NewRouter()
	r.HandleAssets()

	// Middlewares that is common to "web" and "api" endpoint groups
	common := r.Use(
		middleware.Recovery,
		middleware.Logging,
	)

	// Frontend endpoints
	web := common.Group("/")
	web.Use(middleware.ResolveUserFromCookie)
	web.PublicWebEndpoints()
	web.Use(middleware.UserAuthWeb)
	web.ProtectedWebEndpoints()

	// API endpoints
	api := common.Group(conf.ApiBasePath)
	api.Use()

	api.PublicApiEndpoints()

	return &http.Server{
		Addr:         net.JoinHostPort(conf.Host, conf.Port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

func corsConfig() gin.HandlerFunc { // TODO review
	conf := cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Client",
		},
		AllowCredentials: false,
		MaxAge:           24 * time.Hour,
	}

	return cors.New(conf)
}
