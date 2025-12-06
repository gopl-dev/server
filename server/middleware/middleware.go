package middleware

import "github.com/gopl-dev/server/app/service"

type Middleware struct {
	service *service.Service
}

func NewMiddleware(service *service.Service) *Middleware {
	return &Middleware{service: service}
}
