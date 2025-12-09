package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gopl-dev/server/server/endpoint"
)

// Logging is a middleware that logs the start time, HTTP method, and path of an incoming request,
// and then logs the completion time and total duration after the next handler has executed.
func (mw *Middleware) Logging(next endpoint.Handler) endpoint.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Printf("Started %s %s", r.Method, r.URL.Path)
		next(w, r)
		log.Printf("Completed in %v", time.Since(start))
	}
}
