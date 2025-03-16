package api

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

type Handler func(w http.ResponseWriter, r *http.Request)
type Middleware func(Handler) Handler

type Router struct {
	basePath    string
	mux         *http.ServeMux
	routes      map[string]string
	middlewares []Middleware
}

func NewRouter() *Router {
	return &Router{
		basePath:    "/",
		mux:         http.NewServeMux(),
		middlewares: []Middleware{},
	}
}

func (r *Router) Group(pattern string) *Router {
	return &Router{
		basePath:    path.Join(r.basePath, pattern),
		mux:         r.mux,
		middlewares: r.middlewares,
	}
}

func (r *Router) Use(mw ...Middleware) *Router {
	r.middlewares = append(r.middlewares, mw...)

	return r
}

func (r *Router) register(method, pattern string, handler Handler) {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}

	pattern = path.Join(r.basePath, pattern)

	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}
	if !strings.HasSuffix(pattern, "/") {
		pattern = pattern + "/"
	}

	pattern = method + " " + pattern
	fmt.Println(pattern)

	r.mux.Handle(pattern, http.HandlerFunc(handler))
}

func (r *Router) GET(pattern string, handler Handler) *Router {
	r.register(http.MethodGet, pattern, handler)

	return r
}

func (r *Router) POST(pattern string, handler Handler) *Router {
	r.register(http.MethodPost, pattern, handler)

	return r
}

func (r *Router) PUT(pattern string, handler Handler) *Router {
	r.register(http.MethodPut, pattern, handler)

	return r
}

func (r *Router) DELETE(pattern string, handler Handler) *Router {
	r.register(http.MethodDelete, pattern, handler)

	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
