// Package request ...
package request

import "net/http"

// IsJSON checks if the request indicates that it accepts or contains JSON data
// based on the Content-Type header.
func IsJSON(r *http.Request) bool {
	return r.Header.Get("Accept") == "application/json"
}
