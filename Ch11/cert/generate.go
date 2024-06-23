// Pages 256-259
// The command line utility accepts a comma-separated list of hostnames and IP
// addresses that will use the certificate. It also allows you to specify the
// certificate and private-key filenames.
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

var (
	host = flag.String("host", "localhost",
		"Certificate's comma-separated host names and IPs")
	certFn = flag.String("cert", "cert.pem", "certificate file name")
	keyFn  = flag.String("key", "key.pem", "private key file name")
)

// Listing 11-12: Creating an X.509 certificate template.
// The process of generating a certificate and a private key involves building a
// template in your code that you then encode to the X.509 format.
func main() {
	flag.Parse()

	// Each certificate needs a serial number, which a certificate authority
	// typically assigns.
	// Since you're generating your own self-signed certificate, you generate
	// your own serial number using a cryptographically random, unsigned 128-bit integer.
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		log.Fatal(err)
	}

	// You then create an x509.Certificate object that represents an
	// X.509-formatted certificate and sets various values, such as the serial
	// number, the certificate's subject, the validity lifetime, and various
	// usages for this certificate.
	notBefore := time.Now()
	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"Nick Fedor"},
		},
		NotBefore: notBefore,
		NotAfter:  notBefore.Add(10 * 365 * 24 * time.Hour),
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature |
			x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			// Since you want to use this certificate for client authentication, you
			// must include the x509.ExtKeyUsageClientAuth value.
			// If you omit this value, the server won't be able to verify the
			// certificate when presented by the client.
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Listing 11-13:
	// The template is almost ready. You just need to add the hostnames and IP
	// addresses before generating the certificate.
	// You loop through the comma-separated list of hostnames and IP addresses,
	// assigning each to its appropriate slice in the template.
	// Go's TLS client uses these values to authenticate a server.
	// For example, if the client connects to https://www.google.com, but the
	// common name or alternative names in the server's certificate do not match
	// www.google.com's hostname or resolved IP address, then the client fails
	// to authenticate the server.
	for _, h := range strings.Split(*host, ",") {
		if ip := net.ParseIP(h); ip != nil {
			// If the hostname is an IP address, you assign it to the
			// IPAddresses slice.
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			// Otherwise, you assign the hostname to the DNSNames slice.
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// Next, you generate a new ECDSA private key using the P-256 elliptic
	// curve.
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	// At this point, you have everything you need to generate the certificate.
	// The x509.CreateCertificate function accepts a source of entropy
	// (crypto/rand's Reader is ideal), the template for the new certificate, a
	// parent certificate, a public key, and a corresponding private key.
	// It then returns a slice of bytes containing the Distinguished Encoding
	// Rules (DER)-encoded certificate.
	// You use your template for the parent certificate since the resulting
	// certificate signs itself.
	der, err := x509.CreateCertificate(rand.Reader, &template,
		&template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatal(err)
	}

	cert, err := os.Create(*certFn)
	if err != nil {
		log.Fatal(err)
	}

	// All that's left to do is create a new file, generate a new pem.Block with
	// the DEF-encoded byte slice, and PEM-encode everything to the new file.
	err = pem.Encode(cert, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err != nil {
		log.Fatal(err)
	}

	if err := cert.Close(); err != nil {
		log.Fatal(err)
	}

	log.Println("wrote", *certFn)

	// Listing 11-14: Writing the PEM-encoded private key.
	// Now that you have a new certificate on disk, let's write the
	// corresponding private key.
	// Whereas the certificate is meant to be publicly shared, the private key
	// is just that: private.
	// You should take care to assign it minimal permissions.
	// Here, you're giving only the user read-write access to the private-key
	// file and removing access for everyone else.
	key, err := os.OpenFile(
		*keyFn,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal(err)
	}

	// We marshall the private key into a byte slice.
	privKey, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatal(err)
	}

	// We also assign the private key to a new pem.Block before writing the
	// PEM-encoded output to the private-key file.
	err = pem.Encode(
		key,
		&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: privKey})
	if err != nil {
		log.Fatal(err)
	}

	if err := key.Close(); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote", *keyFn)
}
