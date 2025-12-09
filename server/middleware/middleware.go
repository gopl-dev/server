// Package middleware ...
package middleware

import "github.com/gopl-dev/server/app/service"

// Middleware holds dependencies required by middlewares.
type Middleware struct {
	service *service.Service
}

// New creates and returns a new Middleware instance.
func New(service *service.Service) *Middleware {
	return &Middleware{service: service}
}
