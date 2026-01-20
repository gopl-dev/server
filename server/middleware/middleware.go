// Package middleware ...
package middleware

import (
	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/server/handler"
	"go.opentelemetry.io/otel/trace"
)

// Fn is the function signature for a middleware.
type Fn func(handler.Fn) handler.Fn

// Middleware holds dependencies required by middlewares.
type Middleware struct {
	service *service.Service
	tracer  trace.Tracer
}

// New creates and returns a new Middleware instance.
func New(service *service.Service, t trace.Tracer) *Middleware {
	return &Middleware{
		service: service,
		tracer:  t,
	}
}
