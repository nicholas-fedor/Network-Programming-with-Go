// Pages 198-199
// Listing 9-6: Methods map that dynamically routes requests to the right
// handler.
package handlers

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"sort"
	"strings"
)

// Methods is a map whose key is an HTTP method and whose value is an http.Handler.
type Methods map[string]http.Handler

// It has a ServeHTTP method to implement the http.Handler interface, so you can
// use Methods as a handler itself.
func (h Methods) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// The ServeHTTP method first defers a function to drain and close the
	// request body, saving the map's handlers from having to do so.
	defer func(r io.ReadCloser) {
		_, _ = io.Copy(io.Discard, r)
		_ = r.Close()
	}(r.Body)

	if handler, ok := h[r.Method]; ok {
		if handler == nil {
			// The ServeHTTP method makes sure the corresponding handler is not nil,
			// responding with 500 Internal Server Error if it is.
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		} else {
			// Otherwise, you call the corresponding handler's ServeHTTP method.
			// The Methods type is a multiplexer(router) since it routes
			// requests to the appropriate handler.
			handler.ServeHTTP(w, r)
		}

		return
	}

	// If the request method isn't in the map, ServeHTTP responds with the Allow
	// header and a list of supported methods in the map.
	w.Header().Add("Allow", h.allowedMethods())
	// All that's left to do now is determine whether the client explicitly
	// requested the OPTIONS method.
	// If so, the ServeHTTP method returns, resulting in a 200 OK response to
	// the client.
	// If not, the client receives a 405 Method Not Allowed response
	if r.Method != http.MethodOptions {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h Methods) allowedMethods() string {
	a := make([]string, 0, len(h))

	for k := range h {
		a = append(a, k)
	}
	sort.Strings(a)

	return strings.Join(a, ", ")
}

// Pages 199-200
// Listing 9-7: Default implementation of the Methods Handler
func DefaultMethodsHandler() http.Handler {
	return Methods{
		// Now, the handler returned by the handlers.DefaultMethodsHandler
		// function supports the GET, POST, and OPTIONS methods.
		// The GET method simply writes the "Hello, friend!" message to the
		// response body.
		http.MethodGet: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("Hello, friend!"))
			},
		),
		// The POST method greets the client with the HTML-escaped request body
		// contents.
		// The remaining functionality to support the OPTIONS method and
		// properly set the Allow header are inherent to the Methods type's
		// ServeHTTP method.
		http.MethodPost: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				b, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				_, _ = fmt.Fprintf(w, "Hello, %s!", html.EscapeString(string(b)))
			},
		),
	}
}
