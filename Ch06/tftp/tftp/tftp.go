// Page 135
// Listing 6-11: Command line TFTP server implementation.
package main

import (
	"flag"
	"log"
	"os"

	tftp "github.com/nicholas-fedor/Network-Programming-with-Go/Ch06/tftp"
)

var (
	address = flag.String("a", "127.0.0.1:69", "listen address")
	payload = flag.String("p", "payload.svg", "file to serve to clients")
)

func main() {
	flag.Parse()

	p, err := os.ReadFile(*payload)
	if err != nil {
		log.Fatal(err)
	}

	s := tftp.Server{Payload: p}
	log.Fatal(s.ListenAndServe(*address))
}
