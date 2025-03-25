package server

import (
	"net"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gopl-dev/server/app"
)

func NewServer() *http.Server {
	conf := app.Config().Server

	r := NewRouter()
	r.HandleAssets()

	// Middlewares that is common to "web" and "api" endpoint groups
	common := r.Use(
		RecoveryMiddleware,
		LoggingMiddleware,
	)

	// Web endpoints
	web := common.Group("/")
	web.RegisterPublicWebRoutes()

	// API endpoints
	api := common.Group(conf.ApiBasePath)
	api.Use(
		RecoveryMiddleware,
		LoggingMiddleware,
	)

	registerPublicApiRoutes(api)

	return &http.Server{
		Addr:         net.JoinHostPort(conf.Host, conf.Port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

func corsConfig() gin.HandlerFunc {
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
