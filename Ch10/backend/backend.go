// Creating a simple backend web service
// Pages 232-234
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Listing 10-9
// We're setting up a web service that listens on port 8080 of localhost.
// Caddy will direct requests to this socket address.
var addr = flag.String("listen", "localhost:8080", "listen address")

func main() {
	flag.Parse()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Listing 10-10 implements the run function.
	err := run(*addr, c)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server stopped")
}

// Listing 10-10
// The web service receives all requests from Caddy, no matter which client
// originated the request. Likewise, it sends all responses back to Caddy, which
// then routes the response to the right client.
func run(addr string, c chan os.Signal) error {
	mux := http.NewServeMux()
	mux.Handle("/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Conveniently, Caddy adds an X-Forwarded-For header to each
			// request with the originating client's IP address.
			// Although you don't do anything other than log this information,
			// your backend service could use this IP address to differentiate
			// between client requests. For example, the service could deny requests based on
			// the client IP address.
			clientAddr := r.Header.Get("X-Forwarded-For")
			log.Printf("%s -> %s -> %s", clientAddr, r.RemoteAddr, r.URL)
			// The handler writes a slice of bytes to the response that has HTML
			// defined in Listing 10-11.
			_, _ = w.Write(index)
		}),
	)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
	}

	go func() {
		for {
			if <-c == os.Interrupt {
				_ = srv.Close()
				return
			}
		}
	}()

	fmt.Printf("Listening on %s ...\n", srv.Addr)
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		err = nil
	}

	return err
}

var index = []byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Caddy Backend Test</title>
    <link href="/style.css" rel="stylesheet">
</head>
<body>
    <p><img src="/hiking.svg" alt="hiking gopher"></p>
</body>
</html>`)
