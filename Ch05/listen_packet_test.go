// Page 111
// Listing 5-3: Creating an echo server and client.
package echo

import (
	"bytes"
	"context"
	"net"
	"testing"
)

func TestListenPacketUDP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// We start by creating the echo server.
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	// We then create the client connection.
	client, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	// Page 111
	// Listing 5-4: Adding an interloper and interrupting the client with a
	// message.
	// We then create a new UDP connection meant to interlope on the client and
	// echo server and interrupt the client.
	interloper, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	interrupt := []byte("pardon me")
	// This message should queue up in the client's receive buffer.
	n, err := interloper.WriteTo(interrupt, client.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}
	_ = interloper.Close()

	if l := len(interrupt); l != n {
		t.Fatalf("wrote %d bytes of %d", n, l)
	}

	// Page 112
	// Listing 5-5: Receiving UDP packets from multiple senders at once.
	// The client sends its ping message to the echo server and reconciles the
	// replies in Listing 5-5.
	ping := []byte("ping")
	// The client writes a ping message to the echo server.
	_, err = client.WriteTo(ping, serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1024)
	// The client then promptly reads the incoming message.
	n, addr, err := client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	// What's unique about the UDP client connection is it first reads the
	// interruption message from the interloping connection.
	if !bytes.Equal(interrupt, buf[:n]) {
		t.Errorf("expected reply %q; actual reply %q", interrupt, buf[:n])
	}

	if addr.String() != interloper.LocalAddr().String() {
		t.Errorf("expected message from %q; actual sender is %q", interloper.LocalAddr(), addr)
	}

	n, addr, err = client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	// The UDP client connection then reads the reply from the echo server.
	// If this were a TPC connection, then the client would have never received
	// the interloper's message.
	if !bytes.Equal(ping, buf[:n]) {
		t.Errorf("expected reply %q; actual reply %q", ping, buf[:n])
	}

	// As such, we should always verify the sender of each packet it reads by
	// evaluating the second return value (the sender's address) from the ReadFrom method.
	if addr.String() != serverAddr.String() {
		t.Errorf("expected message from %q; actual sender is %q", serverAddr, addr)
	}
}
