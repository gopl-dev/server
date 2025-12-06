package middleware

import (
	"log"
	"net/http"

	"github.com/gopl-dev/server/server/endpoint"
)

func (mw *Middleware) Recovery(next endpoint.Handler) endpoint.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("[%s %s] Recovered from panic: %s", r.Method, r.URL.Path, err)
			}
		}()

		next(w, r)
	}
}
