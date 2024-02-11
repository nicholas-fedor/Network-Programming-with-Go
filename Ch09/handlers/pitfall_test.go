// Pages 196-197
// Listing 9-5: Writing the status first and the response body second for
// expected results.
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerWriteHeader(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// You make a call to the Write method, which implicitly calls
		// WriteHeader(http.StatusOK).
		// Since the the status code is not yet set, the response code is now
		// 200 OK.
		_, _ = w.Write([]byte("Bad request"))
		// The next call to WriteHeader is effectively a no-op because the
		// status code is already set.
		w.WriteHeader(http.StatusBadRequest)
	}
	r := httptest.NewRequest(http.MethodGet, "http://test", nil)
	w := httptest.NewRecorder()
	handler(w, r)
	// The response code 200 Ok persists.
	t.Logf("Response status: %q", w.Result().Status)

	handler = func(w http.ResponseWriter, r *http.Request) {
		// Now, if you switch the order of the calls so you set the status code
		// before you write to the response body ...
		w.WriteHeader(http.StatusBadRequest)
		// ... the response has the proper status code.
		_, _ = w.Write([]byte("Bad request"))
	}

	r = httptest.NewRequest(http.MethodGet, "http://test", nil)
	w = httptest.NewRecorder()
	handler(w, r)
	// 6
	t.Logf("Response status: %q", w.Result().Status)
}
