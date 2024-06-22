// Pages 253-255
// Certificate Pinning
// The process of scrapping the use of the operating system's trusted
// certificate storage and explicitly defining one or more trusted certificates
// in your application.
// Your application will trust connections only from hosts presenting a pinned
// certificate or a certificate signed by a pinned certificate.
// If you plan on deploying clients in zero-trust environments that must
// securely communicate with your server, consider pinning your server's
// certificate to each client.
// Assuming the server introduced in the preceding section uses the cert.pem and
// the key.pem you generated for the hostname localhost, all clients will abort
// the TLS connection as soon as the server presents its certificate.
// Clients wont' trust the server's certificate because no trusted certificate
// authority signed it.
// You could set the tls.Config's InsecureSkipVerify field to true, but as this
// method is insecure, I don't recommend you consider it a practical choice.
// Instead, let's explicitly tell our client it can trust the server's
// certificate by pinning the server's certificate to the client.
package Ch11

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

// Listing 11-9: Creating a new TLS echo server and starting it in the
// background
// Since the hostname in cert.pem is localhost, you create a new TLS echo server
// listening on localhost port 34443.
// The port isn't important here, but clients expect the server to be reachable
// by the same hostname as the one in the certificate it presents.
func TestEchoServerTLS(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverAddress := "localhost:443"
	maxIdle := time.Second
	server := NewTLSServer(ctx, serverAddress, maxIdle, nil)
	done := make(chan struct{})

	go func() {
		// You spin up the server in the background by using the cert.pem and
		// key.pem files.
		err := server.ListenAndServeTLS("cert.pem", "key.pem")
		if err != nil && !strings.Contains(err.Error(),
			"use of closed network connection") {
			t.Error(err)
			return
		}
		done <- struct{}{}
	}()

	// You block until it's ready for incoming connections.
	server.Ready()

	// Listing 11-10: Pinning the server certificate to the client
	// Pinning a server certificate to the client is straightforward.
	// First, you read in the cert.pem file.
	cert, err := ioutil.ReadFile("cert.pem")
	if err != nil {
		t.Fatal(err)
	}

	// Then, you create a new certificate pool and append the certificate to it.
	// As the name suggests, you can add more than one trusted certificate to
	// the certificate pool.
	// This can be useful when you are migrating to a new certificate, but have
	// yet to completely phase out the old certificate.
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(cert); !ok {
		t.Fatal("failed to append certificate to pool")
	}

	// The client, using this configuration, will authenticate only servers that
	// present the cert.pem certificate or any certificate signed to it.
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.CurveP256},
		MinVersion:       tls.VersionTLS12,
		// Finally, you add the certificate pool to the tls.Config's RootCAs field.
		RootCAs: certPool,
	}

	// Listing 11-11: Authenticating the server by using a pinned certificate
	// You pass tls.Dial the tls.Config with the pinned server certificate.
	// Your TLS client authenticates the server's certificate without having to
	// resort to using InsecureSkipVerify and all the insecurity that option introduces.
	conn, err := tls.Dial("tcp", serverAddress, tlsConfig)
	if err != nil {
		t.Fatal(err)
	}

	// Now that you've set up a trusted connection with a server, even though
	// the server presented an unsigned certificate, let's make sure the server
	// works as expected.
	hello := []byte{"hello"}
	_, err = conn.Write(hello)
	if err != nil {
		t.Fatal(err)
	}

	b := make([]byte, 1024)
	n, err := conn.Read(b)
	if err != nil {
		t.Fatal(err)
	}

	// It should echo back any message you sent it.
	if actual := b[:n]; !bytes.Equal(hello, actual) {
		t.Fatalf("expected %q; actual %q", hello, actual)
	}

	// If you idle long enough, you find your next interaction with the socket
	// results in an error, showing the server closed the socket.
	time.Sleep(2 * maxIdle)
	_, err = conn.Read(b)
	if err != io.EOF {
		t.Fatal(err)
	}

	cancel
	<-done
}
