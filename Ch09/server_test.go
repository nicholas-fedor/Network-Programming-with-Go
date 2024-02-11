// Pages 188-189
// Listing 9-1: Instantiating a multiplexer and an HTTP server.
package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
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
			handlers.DefaultHandler(), 2*time.Minute, ""),
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
		{http.MethodPost, bytes.NewBuffer("<world>"), http.StatusOK, "Hello, &lt;world&gt;!"},
		//
		{http.MethodHead, nil, http.StatusMethodNotAllowed, ""},
	}
	client := new(http.Client)
	path := fmt.Sprintf("http://%s/", srv.Addr)
}
