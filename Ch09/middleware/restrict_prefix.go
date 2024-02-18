// Pages 204-205
// Listing 9-13: Protecting any file or directory with a given prefix.
package middleware

import (
	"net/http"
	"path"
	"strings"
)

func RestrictPrefix(prefix string, next http.Handler) http.Handler {
	// The RestrictPrefix middleware...
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// ...examines the URL path to look for any elements that start with
			// a given prefix.
			for _, p := range strings.Split(path.Clean(r.URL.Path), "/") {
				if strings.HasPrefix(p, prefix) {
					// If the middleware finds an element in the URL path with the given
					// prefix, it preempts the http.Handler and response with a 404 Not
					// Found status.
					http.Error(w, "Not Found", http.StatusNotFound)
					return
				}
			}
			next.ServeHTTP(w, r)
		},
	)
}
