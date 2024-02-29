// Pages 211
// Listing 9-18: Command line arguments for the HTTP/2 server.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/nicholas-fedor/Network-Programming-with-Go/Ch09/handlers"
)

var (
	addr = flag.String("listen", "127.0.0.1:8080", "listen address")
	// The server needs the path to a certificate ...
	cert = flag.String("cert", "", "certificate")
	// and a corresponding private key to enable TLS support and allow clients
	// to negotiate HTTP/2 with the server.
	// If either value is empty, the server will listen for plain HTTP connections.
	pkey  = flag.String("key", "", "private key")
	files = flag.String("files", "./files", "static file directory")
)

func main() {
	flag.Parse()

	// Next, pass the command line flag values to a run function.
	// The run function, defined in Listing 9-19, has the bulk of your server's
	// logic and ultimately runs the web server.
	// Breaking this functionality into a separate function eases unit testing later.
	err := run(*addr, *files, *cert, *pkey)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server gracefully shutdown")
}

// Page 212
// Listing 9-19: Multiplexer, middleware, and handlers for HTTP/2 server.
func run(addr, files, cert, pkey string) error {
	mux := http.NewServeMux()
	// The server's multiplexer has three routes: one for static files, ...
	mux.Handle("/static",
		http.StripPrefix(
			".", http.FileServer(http.Dir(files)),
		),
	)
	// ... one for the default route, ...
	mux.Handle("/",
		handlers.Methods{
			http.MethodGet: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					// If the http.ResponseWriter is an http.Pusher, it can push
					// resources to the client ...
					if pusher, ok := w.(http.Pusher); ok {
						targets := []string{
							// You can specify the path to the resource from the
							// client's perspective, not the file path on the
							// server's filesystem because the server treats the
							// request as if the client originated it to
							// facilitate the server push.
							"/static/style.css",
							"/static/hiking.svg",
						}
						for _, target := range targets {
							// ... without a corresponding request.
							if err := pusher.Push(target, nil); err != nil {
								log.Printf("%s push failed: %v", target, err)
							}
						}
					}

					// After you've pushed the resources, you serve hte response
					// for the handler.
					// If, instead, you sent the index.html file before pushing
					// the associated resources, the client's browser may send
					// requests for the associated resources before it handles
					// the pushes.
					http.ServeFile(w, r, filepath.Join(files, "index.html"))
				},
			),
		},
	)
	// ... and one for the /2 absolute path.
	// Web browsers cache HTTP/2-pushed resources for the life of the connection
	// and make it available across routes.
	// Therefore, if the index2.html file served by the /2 route references the
	// same resources pushed by the default route, and the client first visits
	// the default route, the client's web browser may use the pushed resources
	// with rendering the /2 route.
	mux.Handle("/2",
		handlers.Methods{
			http.MethodGet: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					http.ServeFile(w, r, filepath.Join(files, "index2.html"))
				},
			),
		},
	)

	// You have one more task to complete: instantiate an HTTP server to serve
	// your resources.
	// Pages 213-214
	// Listing 9-20: HTTP/2-capable server implementation
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
	}

	done := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		for {
			// When the server receives an os.Interruption signal, ...
			if <-c == os.Interrupt {
				// it triggers a call to the server's Shutdown method.
				// Unlike the server's Close method, which abruptly closes the
				// server's listener and all active connections, Shutdown
				// gracefully shuts down the server.
				// It instructs the server to stop listening for incoming
				// connections and blocks until all client connections end.
				// This gives the server the opportunity to finish sending
				// responses before stopping the server.
				if err := srv.Shutdown(context.Background()); err != nil {
					log.Printf("shutdown: %v", err)
				}
				close(done)
				return
			}
		}
	}()

	log.Printf("Serving files in %q over %s\n", files, srv.Addr)

	var err error
	if cert != "" && pkey != "" {
		log.Println("TLS enabled")
		// If the server receives a path to both the certificate and a
		// corresponding private key, the server will enable TLS support by
		// calling its ListenAndServeTLS method.
		// If it cannot find or parse either the certificate or private key,
		// this method returns an error.
		err = srv.ListenAndServeTLS(cert, pkey)
	} else {
		// In the absence of these paths, the server uses its ListenAndServe method.
		err = srv.ListenAndServe()
	}

	if err == http.ErrServerClosed {
		err = nil
	}

	<-done

	return err
}
