// Pages 188-189
// Listing 9-1: Instantiating a multiplexer and an HTTP server.
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/nicholas-fedor/Network-Programming-with-Go/Ch09/handlers"
)

func TestSimpleHTTPServer(t *testing.T) {
	srv := &http.Server{
		Addr: "127.0.0.1:8081",
		// Requests sent to the server's handler first pass through
		// middleware named http.TimeoutHandler, then to the handler returned by
		// the handlers.DefaultHandler function.
		// In this very simple example, you specify on a single handler for all
		// requests instead of relying on a multiplexer.
		// The server has a few fields.
		// The Handler field accepts a multiplexer or other object capable of
		// handling client requests.
		// The Address field should look familiar by now.
		// In this example, you want the server to listen to port 8081 on IP
		// address 127.0.0.1.
		Handler: http.TimeoutHandler(
			// Updated from handlers.DefaultHandler per Page 200's
			// implementation of Listing 9-7's handlers.DefaultMethodsHandler,
			// which provides additional functionality.
			handlers.DefaultMethodsHandler(), 2*time.Minute, ""),
		IdleTimeout:       5 * time.Minute,
		ReadHeaderTimeout: time.Minute,
	}

	// You create a new net.Listener bound to the server's address ...
	l, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		// ... and instruct the server to Serve requests from this listener.
		// The Serve method returns http.ErrServerClosed when it closes normally.
		err := srv.Serve(l)
		if err != http.ErrServerClosed {
			t.Error(err)
		}
	}()

	// Pages 189-190
	// Listing 9-2: Request test cases for the HTTP server.
	testCases := []struct {
		method   string
		body     io.Reader
		code     int
		response string
	}{
		// First, you send a Get request, which results in a 200 OK status code.
		// The response body has the Hello, friend! string.
		{http.MethodGet, nil, http.StatusOK, "Hello, friend!"},
		// In the second case, you send a POST request with the string <world>
		// in its body.
		// The angle brackets are intentional, and they show an often-overlooked
		// aspect of handling client input in the handler: always escape client
		// input.
		// This test case results in the string Hello, &lt;world&gt;! in the
		// response body.
		{http.MethodPost, bytes.NewBufferString("<world>"), http.StatusOK, "Hello, &lt;world&gt;!"},
		// The third test case sends a HEAD request to the HTTP server.
		// The handler returned by the handlers.DefaultHandler function, which
		// you'll explore shortly, does not handle the HEAD method.
		// Therefore, it returns a 405 Method Not Allowed status code and an
		// empty response body.
		{http.MethodHead, nil, http.StatusMethodNotAllowed, ""},
	}
	client := new(http.Client)
	path := fmt.Sprintf("http://%s/", srv.Addr)

	// Pages 190-191
	// Listing 9-3: Sending test requests to the HTTP server.
	for i, c := range testCases {
		// First, you create a new request, passing the parameters from the test case.
		r, err := http.NewRequest(c.method, path, c.body)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		// Next, you pass the request to the client's Do method, which returns
		// the server's response.
		resp, err := client.Do(r)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		if resp.StatusCode != c.code {
			t.Errorf("%d: unexpected status code: %q", i, resp.Status)
		}

		// You then check the status code and read in the entire response body.
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}
		// You should be in the habit of consistently closing the response body
		// if the client did not return an error, even if the response body is
		// empty or you ignore it entirely.
		// Failure to do so may prevent the client from reusing the underlying
		// TCP connection.
		_ = resp.Body.Close()

		if c.response != string(b) {
			t.Errorf("%d: expected %q; actual %q", i, c.response, b)
		}
	}

	// Once all tests complete, you call the server's Close method.
	// This causes its Serve method in Listing 9-1 to return, stopping the
	// server.
	// The close method abruptly closes client connections.
	if err := srv.Close(); err != nil {
		t.Fatal(err)
	}
}

// Copied from reference code at https://github.com/awoodbeck/gnp/blob/master/ch09/server_test.go
func TestSimpleHTTPServerMethods(t *testing.T) {
	srv := &http.Server{
		Addr: "127.0.0.1:8081",
		Handler: http.TimeoutHandler(
			handlers.DefaultMethodsHandler(), 2*time.Minute, ""),
		IdleTimeout:       5 * time.Minute,
		ReadHeaderTimeout: time.Minute,
	}

	l, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		err := srv.Serve(l)
		if err != http.ErrServerClosed {
			t.Error(err)
		}
	}()

	testCases := []struct {
		method   string
		body     io.Reader
		code     int
		response string
	}{
		{http.MethodGet, nil, http.StatusOK, "Hello, friend!"},
		{http.MethodPost, bytes.NewBufferString("<world>"), http.StatusOK,
			"Hello, &lt;world&gt;!"},
		{http.MethodHead, nil, http.StatusMethodNotAllowed, ""},
	}

	client := new(http.Client)
	path := fmt.Sprintf("http://%s/", srv.Addr)
	for i, c := range testCases {
		r, err := http.NewRequest(c.method, path, c.body)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		resp, err := client.Do(r)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		if resp.StatusCode != c.code {
			t.Errorf("%d: unexpected status code: %q", i, resp.Status)
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}
		_ = resp.Body.Close()

		if c.response != string(b) {
			t.Errorf("%d: expected %q; actual %q", i, c.response, b)
		}
	}

	if err := srv.Close(); err != nil {
		t.Fatal(err)
	}
}
