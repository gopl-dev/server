package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gopl-dev/server/config"
)

func Handler() *gin.Engine {
	conf := config.Get()

	if config.IsReleaseEnv() {
		gin.SetMode(conf.App.Env)
	}

	r := gin.Default()

	r.GET("/api/status/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"env":     conf.App.Env,
			"version": conf.App.Version,
			"time":    time.Now(),
		})
	})

	return r
}
