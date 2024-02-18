// Pages 203-204
// Listing 9-12: Giving clients a finite time to read the response.
package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeoutMiddleware(t *testing.T) {
	// Despite its name, http.TimeoutHandler is middleware that accepts an
	// http.Handler and returns an http.Handler.
	handler := http.TimeoutHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
			// The wrapped http.Handler purposefully sleeps for a minute to
			// simulate a client's taking its time to read a response,
			// preventing the http.Handler from returning.
			time.Sleep(time.Minute)
		}),
		time.Second,
		"Timed out while reading response",
	)

	r := httptest.NewRequest(http.MethodGet, "http://test/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	resp := w.Result()
	// When the handler doesn't return within one second, http.TimeoutHandler
	// sets the response status code to 503 Service Unavailable.
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status code: %q", resp.Status)
	}

	// The test reads the entire response body, properly closes it...
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	// ... and makes sure the response body has the string written by the middleware.
	if actual := string(b); actual != "Timed out while reading response" {
		t.Logf("unexpected body: %q", actual)
	}
}
