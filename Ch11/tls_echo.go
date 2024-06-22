// Pages 249-
package Ch11

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// Listing 11-5: Server struct type and constructor function
// The NewTLSServer function accepts a context for stopping the server, an
// address, a maximum duration the server should allow connections to idle, and
// a TLS configuration.
// Although controlling for idling clients isn't related to TLS, you use the
// maximum idle duration to push the socket deadline forward, as in Chapter 3.
func NewTLSServer(ctx context.Context, address string,
	maxIdle time.Duration, tlsConfig *tls.Config) *Server {
	return &Server{
		ctx:       ctx,
		ready:     make(chan struct{}),
		addr:      address,
		maxIdle:   maxIdle,
		tlsConfig: tlsConfig,
	}
}

// The server struct has a few fields used to record its settings, its TLS
// configuration, and a channel to signal when the server is ready for incoming connections.
type Server struct {
	ctx   context.Context
	ready chan struct{}

	addr      string
	maxIdle   time.Duration
	tlsConfig *tls.Config
}

// You'll write a test case and use the Ready method a little later in this
// section to block until the server is ready to accept connections.
func (s *Server) Ready() {
	if s.ready != nil {
		<-s.ready
	}
}

// Listing 11-6: Adding methods to listen and serve and signal the server's
// readiness for connections.
func (s *Server) ListenAndServeTLS(certFn, keyFn string) error {
	if s.addr == "" {
		s.addr = "localhost:443"
	}

	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("binding to tcp %s: %w", s.addr, err)
	}

	if s.ctx != nil {
		go func() {
			<-s.ctx.Done()
			_ = l.Close()
		}()
	}

	return s.ServeTLS(l, certFn, keyFn)
}

// Listing 11-7: Adding TLS support to a net.Listener
func (s Server) ServeTLS(l net.Listener, certFn, keyFn string) error {
	// The ServeTLS method first checks the server's TLS configuration.
	// If it's nil, it adds a default configuration with
	// PreferServerCipherSuites set to true.
	if s.tlsConfig == nil {
		s.tlsConfig = &tls.Config{
			CurvePreferences: []tls.CurveID{tls.CurveP256},
			MinVersion:       tls.VersionTLS12,
			// PreserveServerCipherSuites is meaningful to the server only, and
			// it makes the server use its preferred cipher suite instead of
			// deferring to the client's preference.
			PreferServerCipherSuites: true,
		}
	}

	if len(s.tlsConfig.Certificates) == 0 &&
		s.tlsConfig.GetCertificate == nil {
		// If the server's TLS configuration does not have at least one certificate,
		// or if its GetCertificate method is nil, you create a new tls.Certificate
		// by reading in the certificate and private-key files from the filesystem.
		cert, err := tls.LoadX509KeyPair(certFn, keyFn)
		if err != nil {
			return fmt.Errorf("loading key pair: %v", err)
		}
		s.tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// At this point in the code, the server has a TLS configuration with at
	// least one certificate ready to present to clients.
	// All that's left is to add TLS support to the net.Listener by passing it
	// and the server's TLS configuration to the tls.NewListener function.
	// The tls.NewListener function acts like middleware, in that it augments
	// the listener to return TLS-aware connection objects from its Accept method.
	tlsListener := tls.NewListener(l, s.tlsConfig)
	if s.ready != nil {
		close(s.ready)
	}

	// Listing 11-8: Accepting TLS-aware connections from the listener
	// This finishes up the ServeTLS method by accepting connections from the
	// listener and handling them in separate goroutines.
	for {
		// You use an endless for loop to continually block on the listener's
		// Accept method, which returns a new net.Conn object when a client
		// successfully connects.
		// Since you're using a TLS-aware listener, it returns connection
		// objects with underlying TLS support.
		// You interact with these connection objects the same as you always do.
		// Go abstracts the TLS details away from you at this point.
		conn, err := tlsListener.Accept()
		if err != nil {
			return fmt.Errorf("accept: %v", err)
		}

		// You then spin off this connection into its own goroutine to handle
		// the connection from that point forward.
		// The server handles each connection the same way.
		go func() {
			defer func() { _ = conn.Close() }()

			for {
				if s.maxIdle > 0 {
					// It firsts conditionally sets the socket deadline to the server's
					// maximum idle duration, then waits for the client to send data.
					err := conn.SetDeadline(time.Now().Add(s.maxIdle))
					if err != nil {
						return
					}
				}

				buf := make([]byte, 1024)
				// If the server doesn't read anything from the socket before it
				// reaches the deadline, then the connection's Read method returns
				// an I/O timeout error, ultimately causing the connection to close.
				n, err := conn.Read(buf)
				if err != nil {
					return
				}

				// If instead, the server reads data from the connection, then
				// it writes that same payload back to the client.
				// Control loops back around to reset the deadline and then wait
				// for the next payload from the client.
				_, err = conn.Write(buf[:n])
				if err != nil {
					return
				}
			}
		}()
	}
}
