package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gopl-dev/server/server/endpoint"
)

func (mw *Middleware) Logging(next endpoint.Handler) endpoint.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		next(w, r)
		log.Printf("Completed in %v", time.Since(start))
	}
}
