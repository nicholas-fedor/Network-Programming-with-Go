// Page 176
// Listing 8-4: The test server causes the default HTTP client to block
// indefinitely.
// If you run this test, use the argument "-timeout 5s" to go test to keep from
// waiting too long.
package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func blockIndefinitely(w http.ResponseWriter, r *http.Request) {
	select {}
}

func TestBlockIndefinitely(t *testing.T) {
	// The net/http/httptest package includes a useful HTTP test server.
	// The httptest.NewServer function accepts an http.HandlerFunc, which in
	// turn wraps the blockIndefinitely function.
	ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))
	// The test server passes any request it receives at its URL to the
	// http.HandlerFunc's ServeHTTP method.
	// This method sends the request and response objects to the
	// blockIndefinitely function, where control blocks indefinitely.
	// Because the helper function http.Get uses the default HTTP client, this
	// GET request won't time out.
	// Instead, the go test runner will eventually time out and halt the test,
	// printing the stack trace.
	_, _ = http.Get(ts.URL)
	t.Fatal("client did not indefinitely block")
}

// Page 177
// Listing 8-5: Adding a time-out to the GET request.
func TestBlockIndefinitelyWithTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First, you create a new request by passing in the context, the request
	// method, the URL, and a nil request body, since your request does not have
	// a payload.
	// 
	// Keep in mind that the context's timer starts running as soon as you
	// initialize the context.
	// The context controls the entire life cycle of the request.
	// In other words, the client has five seconds to connect to the web server,
	// send the request, read the response headers, and pass the response to
	// your code.
	// You then have the remainder of the five seconds to read the response
	// body.
	// 
	// If you are in the middle of reading the response body when the context
	// times out, your next read will immediately return an error.
	// So, use generous time-out values for your specific application.
	// 
	// Alternatively, create a context without a time-out or deadline and
	// control the cancellation of the context exclusively by using a timer and
	// the context's cancel function, like this:
	// ctx, cancel := context.WithCancel(context.Background())
	// timer := time.AfterFunc(5*time.Second, cancel)
	// // Make the HTTP request, read the response headers, etc.
	// // ...
	// // Add 5 more seconds before reading the response body.
	// timer.Reset(5*time.Second)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		return
	}
	_ = resp.Body.Close()
}
