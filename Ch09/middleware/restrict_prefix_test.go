// Pages 205-206
// Listing 9-14: Using the RestrictPrefix middleware.
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRestrictPrefix(t *testing.T) {
	// It's important to realize the server first passes the request to the
	// http.StripPrefix middleware ...
	handler := http.StripPrefix("/static/",
		// ... then the RestrictPrefix middleware...
		RestrictPrefix(".", http.FileServer(http.Dir("../files/"))),
	)

	testCases := []struct {
		path string
		code int
	}{
		// and if the RestrictPrefix middleware approves the resource path, the
		// http.FileServer.
		{"http://test/static/sage.svg", http.StatusOK},
		{"http://test/static/.secret", http.StatusNotFound},
		{"http://test/static/.dir/secret", http.StatusNotFound},
	}

	for i, c := range testCases {
		r := httptest.NewRequest(http.MethodGet, c.path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		actual := w.Result().StatusCode
		if c.code != actual {
			t.Errorf("%d: expected %d; actual %d", i, c.code, actual)
		}
	}
}

// The RestrictPrefix middleware evaluates the request's resource path to
// determine whether the client is requesting a restricted path, no matter
// whether the path exists or not.
// If so, the RestrictPrefix middleware responds to the client with an error
// without ever passing the request onto the http.FileServer.
//
// The static files served by this test's http.FileServer exist in a directory
// named files in the restrict_prefix_test.go file's parent directory.
// Files in the ../files directory are in the root of the filesystem passed to
// the http.FileServer.
// If a client wanted to retrieve the sage.svg file from the http.FileServer,
// the request path should be /sage.svg.
//
// But the URL path for each of our test cases includes the /static/ prefix
// followed by the static filename.
// This means that the test requests static/sage.svg from the http.FileServer,
// which doesn't exist.
// The test uses another bit of middleware from the net/http package to solve
// this path discrepancy.
// The http.StripPrefix middleware strips the given prefix from the URL path
// before passing along the request to the http.Handler, the http.FileServer in
// this test.
//
// Next, you block access to sensitive files by wrapping the http.FileServer
// with the RestrictPrefix middleware to prevent the handler from serving any
// file or directory prefixed with a period.
//
// The first test case results in a 200 OK status, because no element in the URL
// path has a period prefix.
// The http.StripPrefix middleware removes the /static/ prefix from the test
// case's URL, changing it from /static/sage.svg to sage.svg.
// It then passes this path to the http.FileServer, which finds the
// corresponding file in its http.FileSystem.
// The http.FileServer writes the file contents to the response body.
//
// The second test case results in a 404 Not Found status because the .secret
// filename has a period as its first character.
// The third case also results in a 404 Not Found status due to the .dir element
// in the URL path, because your RestrictPrefix middleware considers the prefix
// of each segment in the path, not just the file.
//
// A better approach to restricting access to resources would be to block all
// resources by default and explicitly allow select resources.
