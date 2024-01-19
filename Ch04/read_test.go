// Page 75
// Listing 4-1: Receiving data over a network connection.
package main

import (
	"crypto/rand"
	"io"
	"net"
	"testing"
)

func TestReadIntoBuffer(t *testing.T) {
	// Creates a 16 MB payload of random data that's larger than the client can
	// read in its chosen buffer side of 512 KB, so that it will make at least a
	// few iterations around its for loop.
	payload := make([]byte, 1<<24) // 16 MB
	_, err := rand.Read(payload)   // generate a random payload
	if err != nil {
		t.Fatal(err)
	}

	// Instantiates a listener object.
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	// Creates a goroutine to listen for incoming connections.
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		defer conn.Close()
	}()

	// Once accepted the server rites the entire payload to the network connection.
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	// Creates a 512 KB buffer
	buf := make([]byte, 1<<19) // 512 KB

	// The client reads up ot the first 512 KB from the connection before
	// continuing around the loop.
	// The client continues to read up to 512 KB at a time until either an error
	// occurs or hte client reads the entire 16 MB payload.
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				t.Error(err)
			}
			break
		}

		t.Logf("read %d bytes", n) // buf[:n] is the data read from conn
	}

	conn.Close()
}
