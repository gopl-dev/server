package server

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gopl-dev/server/config"
	"github.com/gopl-dev/server/server/middleware"
)

func Handler() *gin.Engine {
	conf := config.Get()

	if config.IsReleaseEnv() {
		gin.SetMode(conf.App.Env)
	}

	r := gin.Default()
	r.Use(corsConfig())

	api := r.Group(conf.Server.ApiBasePath)
	api.Use(
		middleware.RequestUser,
	)

	r.GET("/api/status/", statusHandler)

	return r
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

func statusHandler(c *gin.Context) {
	conf := config.Get()
	c.JSON(http.StatusOK, gin.H{
		"env":     conf.App.Env,
		"version": conf.App.Version,
		"time":    time.Now(),
	})
}
