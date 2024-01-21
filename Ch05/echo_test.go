// Pages 109-110
// Listing 5-2: Sending UDP packets to the echo server and receiving replies.
package echo

import (
	"bytes"
	"context"
	"net"
	"testing"
)

func TestEchoServerUDP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// We pass a context and the address string to the echoServer function and
	// receive the server's address object.
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	// We defer a call to the context's cancel function, which signals the
	// server to exit and close its goroutines.
	// In a real-world application, using a context for cancellation of
	// long-running processes is useful to make sure we aren't leaking resources
	// like memory or unnecessarily keeping files open.
	defer cancel()

	// We instantiate the client's net.PacketConn in the same way that we
	// instantiated the echo server's net.PacketConn.
	// The net.ListenPacket function creates the connection object for both the
	// client and the server.
	client, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	
	msg := []byte("ping")
	// We need to tell the client where to send its message with each invocation
	// of its WriteTo method.
	_, err = client.WriteTo(msg, serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1024)
	// After sending the message to the echo server, the client should
	// immediately receive a message via its ReadFrom method.
	// We can examine the address returned by the ReadFrom method to confirm
	// that the echo server sent the message.
	n, addr, err := client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	if addr.String() != serverAddr.String() {
		t.Fatalf("received reply from %q instead of %q", addr, serverAddr)
	}

	if !bytes.Equal(msg, buf[:n]) {
		t.Errorf("expected reply %q; actual reply %q", msg, buf[:n])
	}
}
