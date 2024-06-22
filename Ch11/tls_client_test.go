// Pages 245-248
package Ch11

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/http2"
)

// Listing 11-1: Testing HTTPS client and server support
func TestClientTLS(t *testing.T) {
	// The httptest.NewTLSServer function returns an HTTPS server.
	// It handles the HTTPS server's TLS configuration details, including the
	// creation of a new certificate.
	// No trusted authority signed this certificate, so no disconcerting HTTPS
	// client would trust it.
	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				// If the server receives the client's request over HTTP, the
				// request's TLS field will be nil. You can check for this case
				// and redirect the client to the HTTPS endpoint accordingly.
				if r.TLS == nil {
					u := "https://" + r.Host + r.RequestURI
					http.Redirect(w, r, u, http.StatusMovedPermanently)
					return
				}
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer ts.Close()

	// For testing purposes, the server's Client method returns a new
	// *http.Client that inherently trusts the server's certificate.
	// You can use this client to test TLS-specific code within your handlers.
	resp, err := ts.Client().Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d; actual status %d",
			http.StatusOK, resp.StatusCode)
	}

	// Listing 11-2: Testing the HTTPS server with a discerning client.
	// You override the default TLS configuration in your client's
	// transport by creating a new transport, defining its TLS
	// configuration, and configuring http2 to use this transport.
	tp := &http.Transport{
		TLSClientConfig: &tls.Config{
			// It's good practice to restrict your client's curve preference to
			// the P-256 curve and avoid the use of P-384 and P-521.
			// P-256 is immune to timing attacks, whereas P-384 and P-521 are
			// not.
			// Also, your client will negotiate a minimum of TLS 1.2
			// An elliptic curve is a plane curve in which all points along the
			// curve satisfy the same polynomial equation.
			// Whereas first-generation cryptography like RSA uses large prime
			// numbers to derive keys, elliptic curve cryptography uses points
			// along an elliptic curve for key generation.
			// P-256, P-385, and P-521 are the specific elliptic curves defined
			// in the National Institute of Standards and Technology's Digital
			// Signature Standards (FIPS) publication 186-4 (https://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.186-4.pdf)
			CurvePreferences: []tls.CurveID{tls.CurveP256},
			MinVersion:       tls.VersionTLS12,
		},
	}

	// Since your transport no longer relies on the default TLS configuration,
	// the client no longer has inherent HTTP/2 support.
	// You need to explicitly bless your transport with HTTP/2 support if you
	// want to use it. Of course, this test doesn't rely on HTTP/2, but this
	// implementation detail can trip you up if you're unaware that overriding
	// the transport's TLS configuration removes HTTP/2 support.
	err = http2.ConfigureTransport(tp)
	if err != nil {
		t.Fatal(err)
	}

	client2 := &http.Client{Transport: tp}

	// Your client uses the operating system's trusted certificate store because
	// you don't explicitly tell it which certificates to trust.
	// The first call to the test server results in an error because your client
	// doesn't trust the server certificate's signatory.
	_, err = client2.Get(ts.URL)
	if err == nil || !strings.Contains(err.Error(),
		"certificate signed by unknown authority") {
		t.Fatalf("expected unknown authority error; actual %q", err)
	}

	// You could work around this and configure your client's transport to skip
	// verification of the server's certificate by setting its
	// InsecureSkipVerify field to true.
	// This setting is not recommended for anything other than debugging.
	tp.TLSClientConfig.InsecureSkipVerify = true

	// As the field name implies, enabling it makes your client inherently
	// insecure and susceptible to main-in-the-middle attacks, since it now
	// blindly trusts any certificate a server offers up.
	// If you make the same call with your newly naive client, you'll see that
	// it happily negotiates TLS with the server.
	resp, err = client2.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d; actual status %d",
			http.StatusOK, resp.StatusCode)
	}
}

// Listing 11-3: Starting a TLS connection with www.google.com
func TestClientTLSGoogle(t *testing.T) {
	// The tls.DialWithDialer function accepts a *net.Dialer, a network, an
	// address, and a *tls.Config.
	// Here, you give your dialer a timeout of 30 seconds and specify
	// recommended TLS settings.
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 30 * time.Second},
		"tcp",
		"www.google.com:443",
		&tls.Config{
			CurvePreferences: []tls.CurveID{tls.CurveP256},
			MinVersion:       tls.VersionTLS12,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// If successful, you can inspect the connection's state to glean details
	// about your TLS connection.
	state := conn.ConnectionState()
	t.Logf("TLS 1.%d", state.Version-tls.VersionTLS10)
	t.Log(tls.CipherSuiteName(state.CipherSuite))
	t.Log(state.VerifiedChains[0][0].Issuer.Organization[0])

	_ = conn.Close()
}
