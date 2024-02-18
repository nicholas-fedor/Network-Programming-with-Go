// Pages 207-208
// Listing 9-15: Registering patterns to a multiplexer and wrapping the entire
// multiplexer with middleware.
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Here, you use middleware to drain and close the request body.
func drainAndClose(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// You call the "next" handler first and then drain and close the
			// request body.
			next.ServeHTTP(w, r)
			_, _ = io.Copy(io.Discard, r.Body)
			_ = r.Body.Close()
		},
	)
}

// This test creates a new multiplexer and registers three routes using the
// multiplexer's HandleFunc method.
func TestSimpleMux(t *testing.T) {
	serveMux := http.NewServeMux()
	// The first route is simply a forward slash, showing the default or empty
	// URL path, and a function that sets the 204 No Content status in the
	// response.
	// This route will match all URL paths if no other route matches.
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	// The second is "/hello", which writes the string "Hello friend." to the response.
	serveMux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello friend.")
	})
	// The final path is "hello/there/", which writes the string "Why, hello
	// there." to the response.
	//
	// Notice that the third route ends in a forward slash, making it a subtree,
	// while the earlier route did not end in a forward slash, making it an
	// absolute path.
	// This distinction tends to be a bit confusing for unaccustomed users.
	// Go's multiplexer treats absolute paths as exact matches: either the
	// request's URL path matches, or it doesn't.
	// By contrast, it treats subtrees as prefix matches.
	// In other words, the multiplexer will look for the longest registered
	// pattern that comes at the beginning of the request's URL path.
	// For example, "/hello/there" is a prefix of "/hello/there/you" but not of
	// "hello/you".
	//
	// Go's multiplexer can also redirect a URL path that doesn't end in a
	// forward slash, such as "/hello/there".
	// In those cases, the http.ServeMux first attempts to find a matching
	// absolute path.
	// If that fails, the multiplexer appends a forward slash, making the path
	// "/hello/there/", for example, and responds to the client with it.
	// This new path becomes a permanent redirect.
	serveMux.HandleFunc("/hello/there/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Why, hello there.")
	})
	mux := drainAndClose(serveMux)

	// Pages 208-209
	// Listing 9-16: Running through a series of test cases and verifying the
	// response status code and body.
	testCases := []struct {
		path     string
		response string
		code     int
	}{
		// The first three test cases, including the request for the "/hello/there/"
		// path, match exact patterns registered with the multiplexer.
		{"http://test/", "", http.StatusNoContent},
		{"http://test/hello", "Hello friend.", http.StatusOK},
		{"http://test/hello/there/", "Why, hello there.", http.StatusOK},
		// But the fourth test case is different. It doesn't have an exact
		// match. When the multiplexer appends a forward slash to it, however,
		// it discovers that it exactly matches a registered pattern. Therefore,
		// the multiplexer responds with a 301 Moved Permanently status and a
		// link to the new path in the response body.
		{"http://test/hello/there", "<a href=\"/hello/there/\">Moved Permanently</a>.\n\n", http.StatusMovedPermanently},
		// The fifth response matches the "/hello/there/" subtree and receives
		// the "Why, hello there." response.
		{"http://test/hello/there/you", "Why, hello there.", http.StatusOK},
		// The last three test cases match the default path of "/" and receive
		// the 204 No Content status.
		{"http://test/hello/and/goodbye", "", http.StatusNoContent},
		{"http://test/something/else/entirely", "", http.StatusNoContent},
		{"http://test/hello/you", "", http.StatusNoContent},
	}

	for i, c := range testCases {
		r := httptest.NewRequest(http.MethodGet, c.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		resp := w.Result()

		if actual := resp.StatusCode; c.code != actual {
			t.Errorf("%d: expected code %d; actual %d", i, c.code, actual)
		}

		// Just as the test relies on middleware to drain and close the request
		// body, it drains...
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		// ...and closes the response body.
		_ = resp.Body.Close()

		if actual := string(b); c.response != actual {
			t.Errorf("%d: expected response %q; actual %q", i, c.response, actual)
		}
	}
}
