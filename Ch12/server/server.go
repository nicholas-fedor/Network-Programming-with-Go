// Pages 288-289
// Listing 12-22: Creating a new gRPC server using Rosie
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"Ch12/housework/v1"
)

var addr, certFn, keyFn string

func init() {
	flag.StringVar(&addr, "address", "localhost:34443", "listen address")
	flag.StringVar(&certFn, "cert", "cert.pem", "certificate file")
	flag.StringVar(&keyFn, "key", "key.pem", "private key file")
}

func main() {
	flag.Parse()

	// First, you retrieve a new server instance.
	server := grpc.NewServer()
	rosie := new(Rosie)
	// You pass it and a new *housework.RobotMaidService from Rosie's Service
	// method to the RegisterRobotMaidServer function in the generated gRPC
	// code.
	// This registers Rosie's RobotMaidService implementation with the gRPC server.
	housework.RegisterRobotMaidService(server, rosie.Service())

	cert, err := tls.LoadX509KeyPair(certFn, keyFn)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Listening for TLS connections on %s ...", addr)
	// You call the server's Serve method.
	// You then load the server's key pair and create a new TLS
	// listener, which you pass to the server when calling Serve.
	log.Fatal(server.Serve(tls.NewListener(listener,
		&tls.Config{
			Certificates:             []tls.Certificate{cert},
			CurvePreferences:         []tls.CurveID{tls.CurveP256},
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
		},
	)))
}
