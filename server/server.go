// Package server ...
package server

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/server/endpoint"
	"github.com/gopl-dev/server/server/handler"
	"github.com/gopl-dev/server/server/middleware"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"go.opentelemetry.io/otel/trace"
)

// RWTimeout defines server's Read&Write timeout in seconds.
const RWTimeout = 10 * time.Second

// New creates new server.
func New(s *service.Service, t trace.Tracer) *http.Server {
	conf := app.Config().Server

	registerOAuthProviders()

	h := handler.New(s, t)
	mw := middleware.New(s, t)
	r := endpoint.NewRouter(mw, h)

	r.HandleAssets()
	r.HandleOpenAPIDocs()

	// Middlewares that is common to "web" and "api" endpoint groups
	common := r.Use(
		mw.Tracing,
		mw.Recovery,
		mw.Logging,
		mw.ResolveUserFromCookie,
	)

	// Frontend endpoints
	web := common.Group("/")
	web.PublicWebEndpoints()
	web.Use(mw.UserAuth)
	web.ProtectedWebEndpoints()
	common.HandleNotFound()

	// API endpoints
	api := common.Group(conf.APIBasePath)
	api.Use()

	api.PublicAPIEndpoints()
	api.Use(mw.UserAuth)
	api.ProtectedAPIEndpoints()
	api.HandleNotFound()

	return &http.Server{
		Addr:         net.JoinHostPort(conf.Host, conf.Port),
		Handler:      r,
		ReadTimeout:  RWTimeout,
		WriteTimeout: RWTimeout,
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
		// MaxAge:           24 * time.Hour,
	}

	return cors.New(conf)
}

func registerOAuthProviders() {
	c := app.Config()

	callbackURL := func(providerName string) string {
		return fmt.Sprintf("%sauth/%s/callback/", c.Server.Addr, providerName)
	}

	goth.UseProviders(
		google.New(c.GoogleOAuth.ClientID, c.GoogleOAuth.ClientSecret, callbackURL("google")),
		github.New(c.GithubOAuth.ClientID, c.GithubOAuth.ClientSecret, callbackURL("github")),
	)
}
