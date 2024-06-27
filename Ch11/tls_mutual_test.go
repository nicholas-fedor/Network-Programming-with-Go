// Pages 260-265
package Ch11

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"os"
	"strings"
	"testing"
)

// Listing 11-16: Creating a certificate pool to serve CA certificates
// Both the client and server use the caCertPool function to create a new X.509
// certificate pool.
// The certificate pool serves as a source of trusted certificates. The client
// puts the server's certificate in its certificate pool, and vice versa.
func caCertPool(caCertFn string) (*x509.CertPool, error) {
	// The function accepts the file path to a PEM-encoded certificate, which
	// you read in...
	// Note: Code from book references ioutil.ReadFile; however, this is deprecated.
	// Using os.ReadFile instead.
	caCert, err := os.ReadFile(caCertFn)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	// ... and append to the new certificate pool.
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return nil, errors.New("failed to add certificate to pool")
	}

	return certPool, nil
}

// Listing 11-17: Instantiating a CA cert pool and a server certificate.
// TestMutualTLSAuthentication details the initial test code to demonstrate
// mutual TLS authentication between a client and a server.
func TestMutualTLSAuthentication(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Before creating the server, you need to first populate a new CA
	// certificate pool with the client's certificate.
	serverPool, err := caCertPool("clientCert.pem")
	if err != nil {
		t.Fatal(err)
	}

	// You also need to load the server's certificate at this point instead of
	// relying on the server's ServeTLS method to do it for you.
	cert, err := tls.LoadX509KeyPair("serverCert.pem", "serverKey.pem")
	if err != nil {
		t.Fatalf("loading key pair: %v", err)
	}

	// Listing 11-18: Accessing the client's hello information using
	// GetConfigForClient
	// Recall that in Listing 11-13, you defined the IPAddresses and DNSNames
	// slices of the template used to generate your client's certificate.
	// These values populate the common name and alternative names portions of
	// the client's certificate.
	// You learned that Go's TLS client uses these values to authenticate the
	// server. But the server does not use these values from the client's certificate to
	// authenticate the client.
	// Since you're implementing mutual TLS authentication, you need to make
	// some changes to the server's certificate verification process so that it
	// authenticates the client's IP address or hostnames against the client
	// certificate's common name and alternative names.
	// To do that, the server at the very least needs to know the client's IP address.
	serverConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// The only way you can get client connection information before
		// certificate verification is by defining the tls.Config's
		// GetConfigForClient method.
		// This method allows you to define a function that receives the
		// *tls.ClientHelloInfo object created as part of the TLS handshake
		// process with the client.
		// From this, you can retrieve the client's IP address.
		// But first, you need to return a proper TLS configuration.
		// This function returns the same TLS configuration for every client
		// connection.
		// As mentioned, the only reason you're using the GetConfigForClient
		// method is so you can retrieve the client's IP from its hello information.
		GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			return &tls.Config{
				// You add the server's certificate to the TLS configuration
				Certificates: []tls.Certificate{cert},
				// You also need to tell the server that every client must
				// present a valid certificate before completing the TLS
				// handshake process.
				ClientAuth: tls.RequireAndVerifyClientCert,
				// You also add the server pool to the TLS configuration's
				// ClientCA's field.
				// This field is the server's equivalent to the TLS
				// configuration's RootCAs field on the client.
				ClientCAs:        serverPool,
				CurvePreferences: []tls.CurveID{tls.CurveP256},
				// Since you control both the client and server, specify a
				// minimum TLS protocol version of 1.3.
				MinVersion:               tls.VersionTLS13,
				PreferServerCipherSuites: true,

				// Listing 11-19: Making the server authenticate the client's IP
				// and hostnames
				// Implements the verification process that authenticates the
				// client by using its IP address and its certificate's common
				// name and alternative names.
				// Since you want to augment the usual certificate verification
				// process on the server, you define and appropriate function
				// and assign it to the TLS configuration's
				// VerifyPeerCertificate method.
				// The server calls this method after the normal certificate
				// verification checks.
				// The only check you're performing above and beyond the normal
				// checks is to verify the client's hostname with the leaf
				// certificate.
				// The leaf certificate is the last certificate in the
				// certificate chain given to the server by the client.
				// The leaf certificate contains the client's public key.
				// All other certificates in the chain are intermediate
				// certificates used to verify the authenticity of the leaf
				// certificate and culminate with the certificate authority's
				// certificate.
				// You'll find each leaf certificate at index 0 in each
				// verifiedChains slice.
				// In other words, you can find the leaf certificate of the
				// first chain at verifiedChains[0][0].
				// If the server calls your function assigned to the
				// VerifyPeerCertificate method, the leaf certificate in the
				// first chain exists at a minimum.
				VerifyPeerCertificate: func(rawCerts [][]byte,
					verifiedChains [][]*x509.Certificate) error {
					opts := x509.VerifyOptions{
						// Create a new x509.VerifyOptions object
						KeyUsages: []x509.ExtKeyUsage{
							// Modify the KeyUsages method to indicate you want
							// to perform client authentication.
							x509.ExtKeyUsageClientAuth,
						},
						// Then, assign the server pool to the Roots method.
						// The server uses this pool as its trusted certificate
						// source during verification.
						Roots: serverPool,
					}

					// Now, extract the client's IP address from the connection
					// object in the *tls.ClientHelloInfo object named hello
					// passed into Listing 11-18's GetConfigForClient method.
					ip := strings.Split(hello.Conn.RemoteAddr().String(),
						":")[0]
					// Use the IP address to perform a reverse DNS lookup to
					// consider any hostnames assigned to the client's IP
					// address.
					hostnames, err := net.LookupAddr(ip)
					// If this lookup fails or returns an empty slice, then the
					// way you handle that situation is up to you.
					// If you're relying on the client's hostname for
					// authentication and the reverse lookup fails, then you
					// cannot authenticate the client.
					// But, if you're using the client's IP address only in the
					// certificate's common name or alternative names, then a
					// reverse lookup failure is inconsequential.
					// For demonstration purposes, we'll consider a failed
					// reverse lookup to equate to a failed test.
					if err != nil {
						t.Errorf("PTR lookup: %v", err)
					}
					// At minimum, you append the client's IP address to the
					// hostnames slice.
					hostnames = append(hostnames, ip)

					// All that's left to do is loop through each verified
					// chain.
					for _, chain := range verifiedChains {
						// Assign a new intermediate certificate pool
						// to opts.Intermediates.
						opts.Intermediates = x509.NewCertPool()

						// Add all certificates but the leaf certificates to the
						// intermediate certificate pool.
						for _, cert := range chain[1:] {
							opts.Intermediates.AddCert(cert)
						}

						for _, hostname := range hostnames {
							// And attempt to verify the client.
							opts.DNSName = hostname
							_, err = chain[0].Verify(opts)
							// If the verification returns a nil error, then you
							// authenticated the client.
							if err == nil {
								return nil
							}
						}
					}

					// If you fail to verify each hostname with each leaf
					// certificate, then return an error to indicate that client
					// authentication failed.
					// The client will receive an error and the server will
					// terminate the connection.
					return errors.New("client authentication failed")
				},
			}, nil
		},
	}

	// Listing 11-20: Starting the TLS server
	// Now that the server's TLS configuration properly authenticates client
	// certificates, continue with the server implementation.
	serverAddress := "localhost:44443"
	// Create a new TLS server instance, making sure to pass in the TLS
	// configuration you just created.
	server := NewTLSServer(ctx, serverAddress, 0, serverConfig)
	done := make(chan struct{})

	// Call its ListenAndServeTLS method in a goroutine
	go func() {
		err := server.ListenAndServeTLS("serverCert.pem", "serverKey.pem")
		if err != nil && !strings.Contains(err.Error(),
			"use of closed network connection") {
			t.Error(err)
			return
		}
		done <- struct{}{}
	}()
	// Make sure to wait until the server is ready for connections before proceeding.
	server.Ready()

	// Listing 11-21: Pinning the server certificate to the client
	// Now that the server implementation is ready, let's move onto the client
	// portion of the test.
	// This implements a TLS client that can present clientCert.pem upon request
	// by the server.
	// The client retrieves a new certificate pool populated with the server's certificate.
	clientPool, err := caCertPool("serverCert.pem")
	if err != nil {
		t.Fatal(err)
	}

	clientCert, err := tls.LoadX509KeyPair("clientCert.pem", "clientKey.pem")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := tls.Dial("tcp", serverAddress, &tls.Config{
		// You also configure the client with its own certificate to present to
		// the server upon request.
		Certificates:     []tls.Certificate{clientCert},
		CurvePreferences: []tls.CurveID{tls.CurveP256},
		MinVersion:       tls.VersionTLS13,
		// The client then uses the certificate pool in the RootCA's field of
		// its TLS configuration, meaning the client will trust only server
		// certificates signed by serverCert.pem.
		RootCAs: clientPool,
	})
	if err != nil {
		t.Fatal(err)
	}
	// It's worth noting that the client and server have not initialized a TLS
	// session yet. They haven't completed the TLS handshake.
	// If tls.Dial returns an error, it isn't because of an authentication
	// issue, but more likely a TCP connection issue.

	// Listing 11-22: TLS handshake completes as you interact with the
	// connection
	hello := []byte("hello")
	_, err = conn.Write(hello)
	if err != nil {
		t.Fatal(err)
	}

	b := make([]byte, 1024)
	// The first read from, or write to, the socket connection automatically
	// initializes the handshake process between the client and server.
	// If the server rejects the client certificate, then the read call will
	// return a bad certificate error.
	// But if you created appropriate certificates and properly pinned them,
	// then both the client and the server are happy, and this test passes.
	n, err := conn.Read(b)
	if err != nil {
		t.Fatal(err)
	}

	if actual := b[:n]; !bytes.Equal(hello, actual) {
		t.Fatalf("expected %q; actual %q", hello, actual)
	}

	err = conn.Close()
	if err != nil {
		t.Fatal(err)
	}

	cancel()
	<-done
}
