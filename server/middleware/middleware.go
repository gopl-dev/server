// Package middleware ...
package middleware

import (
	"github.com/gopl-dev/server/app/service"
	"go.opentelemetry.io/otel/trace"
)

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
