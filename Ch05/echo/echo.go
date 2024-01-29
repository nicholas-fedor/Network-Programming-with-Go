// Pages 107-108
// Listing 5-1: A simple UDP echo server implementation.
package echo

import (
	"context"
	"fmt"
	"net"
)

// echoServerUDP receives a context to allow cancellation of the echo server by
// the caller and a string address in the host:port format.
// It returns a net.Addr interface and an error interface.
// The caller uses the net.Addr interface to address messages to the echo
// server.
// The returned error interface is not nil if anything goes wrong while
// instantiating the echo server.
func echoServerUDP(ctx context.Context, addr string) (net.Addr, error) {

	// Creates a UDP connection for the server with a call to net.ListenPacket,
	// which returns a net.PacketConn interface and an error interface.
	// The net.ListenPacket function is analogous to the net.Listen function
	// used to create a TCP listener, except net.ListenPacket exclusively
	// returns a net.PacketConn interface.
	s, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("binding to udp %s: %w", addr, err)
	}

	// The goroutine manages the asynchronous echoing of messages by the echo server.
	go func() {
		go func() {
			// A second goroutine blocks on the context's Done channel.
			// Once the caller cancels the context, receiving on the Done channel
			// unblocks and the server is closed, tearing down both this goroutine
			// and the parent goroutine.
			<-ctx.Done()
			_ = s.Close()

		}()

		buf := make([]byte, 1024)

		for {
			// To read from the UDP connection, we pass a byte slice to the
			// ReadFrom method.
			// This returns the number of bytes read, the sender's address, and
			// an error interface.
			// Notice there is no Accept method on your UDP connection as there
			// is with TCP-based listeners.
			// This is because UDP doesn't use a handshake process.
			// Here, we simply create a UDP connection listening to a UDP port
			// and read any incoming messages.
			// Since we don't have a proper introduction and an established
			// session, we rely on the returned address to determine which node
			// sent us the message.
			n, clientAddr, err := s.ReadFrom(buf) // client to server
			if err != nil {
				return
			}

			// To write a UDP packet, we pass a byte slice and a destination
			// address to the connection's WriteTo method.
			// The WriteTo method returns the number of bytes written and an
			// error interface.
			// Just as with reading data, we need to tell the WriteTo method
			// where to send the packet, because we do not have an established
			// session with a remote node.
			// In Listing 5-1, we write the message to the original sender.
			// But we could just as easily forward the message to another node
			// using our existing UDP connection object.
			// We do not have to establish a new UDP connection object to
			// forward on the message as we would if using TCP.
			_, err = s.WriteTo(buf[:n], clientAddr) // server to client
			if err != nil {
				return
			}
		}
	}()

	return s.LocalAddr(), nil
}
