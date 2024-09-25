// Pages 320-
package main

// Pages 320-321
// Listing 13-24: Imports and command line flags for the metrics example
import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	// The only imports your code needs are the promhttp package for the metrics
	// endpoint and your umetrics package to instrument your code.
	// The promhttp package includes an http.Handler that a Prometheus server
	// can use to scrap metrics from your application.
	// This handler serves not only your metrics but also metrics related to the
	// runtime, such as the Go version, number of cores, and so on.
	// At a minimum, you can use the metrics provided by the Prometheus handler
	// to gain insight into your service's memory utilization, open file
	// descriptors, heap and stack details, and more.
	"github.com/prometheus/client_golang/prometheus/promhttp"
	// All variable exported by your metrics package are Go kit interfaces.
	"Ch13/instrumentation/metrics"
)

var (
	metricsAddr = flag.String("metrics", "127.0.0.1:8081", "metrics listen address")
	webAddr     = flag.String("web", "127.0.0.1:8082", "web listen address")
)

// Pages 321-322
// Listing 13-25: An instrumented handler that responds with random latency.
// Even in such a simple handler, you're able to make three meaningful
// measurements.
// You increment the requests counter upon entering the handler since it's the
// most logical place to account for it.
// You also immediately defer a function that calculates the request duration
// and uses the request duration summary metric to observe it.
// Lastly, you account for any errors writing the response.
func helloHandler(w http.ResponseWriter, _ *http.Request) {
	metrics.Requests.Add(1)
	defer func(start time.Time) {
		metrics.RequestDuration.Observe(time.Since(start).Seconds())
	}(time.Now())
	_, err := w.Write([]byte("Hello!"))
	if err != nil {
		metrics.WriteErrors.Add(1)
	}
}

// Page 322
// Listing 13-26: Functions to create an HTTP server and instrument connection
// states.
func newHTTPServer(addr string, mux http.Handler,
	// This HTTP server code resembles that of Chapter 9.
	// The exception here is you're defining the server's ConnState field,
	// accepting it as an argument to the newHTTPServer function.
	// The HTTP server calls its ConnState field anytime a network connection
	// changes.
	// You can leverage this functionality to instrument the number of open
	// connections the server has at any one time.
	stateFunc func(net.Conn, http.ConnState)) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
		ConnState:         stateFunc,
	}

	go func() { log.Fatal(srv.Serve(l)) }()

	return nil
}

// You can pass the connStateMetrics function to the newHTTPServer function
// anytime you want to initialize a new HTTP server and track its open connections.
func connStateMetrics(_ net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		// If the server establishes a new connection, you increment the open
		// connections gauge by 1.
		metrics.OpenConnections.Add(1)
	case http.StateClosed:
		// If a connection closes, you decrement the gauge by 1.
		// Go kit's gauge interface provides an Add method, so decrementing a
		// value involves adding a negative number.
		metrics.OpenConnections.Add(-1)
	}
}

// Page 323
// Listing 13-27: Starting two HTTP servers to serve metrics and the helloHandler.
func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	mux := http.NewServeMux()
	// First, you spawn an HTTP server with the sole purpose of serving the
	// Prometheus handler at the /metrics/ endpoint where Prometheus scrapes
	// metrics from by default.
	mux.Handle("/metrics", promhttp.Handler())
	// Since you do not pass in a function for the third argument, this HTTP
	// server won't have a function assigned to its ConnState field to call on
	// each connection state change.
	if err := newHTTPServer(*metricsAddr, mux, nil); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Metrics listening on %q ...\n", *metricsAddr)

	// Then, you spin up another HTTP server to handle each request with the
	// helloHandler.
	// But this time, you pass in the connStateMetrics function.
	// As a result, this HTTP server wil gauge open connections.
	if err := newHTTPServer(*webAddr, http.HandlerFunc(helloHandler),
		connStateMetrics); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Web listening on %q ...\n\n", *webAddr)

	// Page 324
	// Listing 13-28: Instructing 500 HTTP clients to each make 100 GET calls

	// You start by spawning 500 clients to each make 100 GET calls.
	clients := 500
	gets := 100

	wg := new(sync.WaitGroup)

	fmt.Printf("Spawning %d connections to make %d requests each ...\n", clients, gets)
	for i := 0; i < clients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			c := &http.Client{
				// The http.Client uses the http.DefaultTransport if its
				// Transport method is nil.
				// It does an outstanding job of caching TCP connections.
				// If all 500 HTTP clients use the same transport, they'll all
				// make calls over about two TCP sockets.
				// Our open connections gauge would reflect the idle connections
				// when you're done with this example, which really isn't the
				// goal.
				// Instead, you much make sure to give each HTTP client its own
				// transport.
				// Cloning the default transport is good enough for our purposes.
				Transport: http.DefaultTransport.(*http.Transport).Clone(),
			}

			for j := 0; j < gets; j++ {
				// Now that each client has its own transport and you're assured
				// each client will make its own TCP connection, you iterate
				// through a GET call 100 times with each client.
				resp, err := c.Get(fmt.Sprintf("http://%s/", *webAddr))
				if err != nil {
					log.Fatal(err)
				}
				// You must also be diligent about draining and closing the
				// response body so each client can reuse its TCP connection.
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
			}
		}()
	}
	// Once all 500 HTTP clients complete their 100 calls, you can move on to
	// check the current state of the metrics.
	wg.Wait()
	fmt.Print(" done.\n\n")

	// Page 325
	// Listing 13-29: Displaying the current metrics matching your namespace and
	// subsystem.
	
	// You retrieve all the metrics from the metrics endpoint.
	// This will cause the metrics web server to return all metrics stored by
	// the Prometheus client, in addition to details about each metric it
	// tracks, which includes the metrics you added.
	resp, err := http.Get(fmt.Sprintf("http://%s/metrics", *metricsAddr))
	if err != nil {
		log.Fatal(err)
	}
	
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	_ = resp.Body.Close()
	
	// Since you're only interested in your metrics, you can check each line
	// starting with your namespace, an underscore, and your subsystem.
	metricsPrefix := fmt.Sprintf("%s_%s", *metrics.Namespace, *metrics.Subsystem)
	fmt.Println("Current Metrics:")
	for _, line := range bytes.Split(b, []byte("\n")) {
		// If the line matches this prefix, you print it to standard output.
		// Otherwise, you ignore the line and move on.
		if bytes.HasPrefix(line, []byte(metricsPrefix)) {
			fmt.Printf("%s\n", line)
		}
	}
}
