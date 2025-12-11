// Package endpoint ...
package endpoint

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/frontend"
	"github.com/gopl-dev/server/server/docs"
	"github.com/gopl-dev/server/server/handler"
	httpSwagger "github.com/swaggo/http-swagger"
)

const exactMatchSuffix = "{$}"

// Handler is the function signature for a standard request handler.
type Handler func(w http.ResponseWriter, r *http.Request)

// Middleware is the function signature for a middleware.
type Middleware func(Handler) Handler

// Router manages the application's routing table, handles middleware application,
// and delegates requests to the standard http.ServeMux.
type Router struct {
	basePath    string
	mux         *http.ServeMux
	routes      map[string]string
	handler     *handler.Handler
	middlewares []Middleware
}

// NewRouter initializes and returns a new Router instance.
func NewRouter(h *handler.Handler) *Router {
	return &Router{
		basePath:    "/",
		mux:         http.NewServeMux(),
		handler:     h,
		middlewares: []Middleware{},
	}
}

// Group creates a new Router instance with a base path appended to the current router's base path.
// The new router shares the underlying http.ServeMux and middleware stack.
func (r *Router) Group(pattern string) *Router {
	return &Router{
		basePath:    path.Join(r.basePath, pattern),
		mux:         r.mux,
		handler:     r.handler,
		middlewares: r.middlewares,
	}
}

// Use appends one or more Middleware functions to the router's middleware stack.
func (r *Router) Use(mw ...Middleware) *Router {
	r.middlewares = append(r.middlewares, mw...)

	return r
}

// GET registers a handler for the HTTP GET method at the specified pattern relative to the base path.
func (r *Router) GET(pattern string, handler Handler) *Router {
	r.register(http.MethodGet, pattern, handler)

	return r
}

// POST registers a handler for the HTTP POST method at the specified pattern relative to the base path.
func (r *Router) POST(pattern string, handler Handler) *Router {
	r.register(http.MethodPost, pattern, handler)

	return r
}

// PUT registers a handler for the HTTP PUT method at the specified pattern relative to the base path.
func (r *Router) PUT(pattern string, handler Handler) *Router {
	r.register(http.MethodPut, pattern, handler)

	return r
}

// DELETE registers a handler for the HTTP DELETE method at the specified pattern relative to the base path.
func (r *Router) DELETE(pattern string, handler Handler) *Router {
	r.register(http.MethodDelete, pattern, handler)

	return r
}

// HandleAssets registers a file server handler to serve static assets from the "/assets/" path.
func (r *Router) HandleAssets() *Router {
	r.mux.Handle("GET /assets/", http.FileServer(http.FS(frontend.AssetsFs)))
	return r
}

// HandleOpenAPI registers a file server handler to serve static swagger files.
func (r *Router) HandleOpenAPI() *Router {
	conf := app.Config()
	if !conf.OpenAPI.Enabled {
		return r
	}

	serverURL, err := url.Parse(conf.Server.Addr)
	if err != nil {
		log.Printf("OpenAPIHandler: could not parse server address: %s\n", err)
		return r
	}

	docs.SwaggerInfo.Title = fmt.Sprintf("%s API (%s %s)", conf.App.Name, conf.App.Env, conf.App.Version)
	docs.SwaggerInfo.Host = serverURL.Host
	docs.SwaggerInfo.BasePath = "/" + conf.Server.APIBasePath
	docs.SwaggerInfo.Schemes = []string{serverURL.Scheme}

	servePath := "/" + strings.Trim(conf.OpenAPI.ServePath, "/") + "/"
	docPath := servePath + "doc.json"

	r.mux.Handle(servePath, httpSwagger.Handler(
		httpSwagger.URL(docPath),
	))

	return r
}

// HandleNotFound registers a handler for any request that does not match other registered routes
// within the current router's base path.
func (r *Router) HandleNotFound() *Router {
	pattern := r.basePath

	if !strings.HasSuffix(pattern, "/") {
		pattern += "/"
	}

	r.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		println(r.Method, r.RequestURI, "404 NOT FOUND")
		http.NotFound(w, r)
	})

	return r
}

// ServeHTTP implements the http.Handler interface, delegating the request handling
// to the underlying http.ServeMux.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) register(method, pattern string, handler Handler) {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}

	withSlash := strings.HasSuffix(pattern, "/")
	pattern = path.Join(r.basePath, pattern)

	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	if withSlash && !strings.HasSuffix(pattern, "/") {
		pattern += "/" + exactMatchSuffix
	}

	if pattern == "/" {
		pattern += exactMatchSuffix
	}

	pattern = method + " " + pattern
	fmt.Println(pattern)

	r.mux.Handle(pattern, http.HandlerFunc(handler))
}
