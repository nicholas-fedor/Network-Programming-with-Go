// Page 113
// Listing 5-6: Creating an echo server and client.
// Creates the UDP-based net.Conn and demonstrates how net.Conn encapsulates the
// implementation details of UDP to emulate a stream-oriented network connection.
// The client side of a connection can leverage the stream-oriented
// functionality of net.Conn over UDP, but the UDP listener must still use net.PacketConn.
package echo

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"
)

func TestDialUDP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// We span an instance of the echo server for the purpose of sending a reply
	// to the client.
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	// We then dial the echo server over UDP by passing udp as the first
	// argument to net.Dial.
	// Unlike TCP, the echo server receives no traffic upon calling net.Dial
	// because no handshake is necessary.
	client, err := net.Dial("udp", serverAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	// Pages 113-114
	// Listing 5-7: Interrupting the client.
	// This interrupts the client by sending a message to it before the echo
	// server sends its reply.
	interloper, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	interrupt := []byte("pardon me")
	// Sends a message to the client from an interloping connection.
	n, err := interloper.WriteTo(interrupt, client.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}
	_ = interloper.Close()

	if l := len(interrupt); l != n {
		t.Fatalf("wrote %d bytes of %d", n, l)
	}

	// Page 114
	// Listing 5-8: Using net.Conn to manage UPD traffic.
	// This details the difference between a UDP connection using net.Conn and
	// one using net.PacketConn.
	ping := []byte("ping")
	// The client sends a ping message to the echo server by using net.Conn's
	// Write method.
	// The net.Conn client will write its messages to the address specified in
	// the net.Dial call.
	// We do not need to specify a destination address for every packet we send
	// using hte client connection.
	_, err = client.Write(ping)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1024)
	// Likewise, we read packets using the client's Read method.
	// The client reads packets only from the sender address specified in the
	// net.Dial call, as we would expect using a stream-oriented connection
	// object.
	// The client never reads the message sent by the interloping connection.
	n, err = client.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(ping, buf[:n]) {
		t.Errorf("expected reply %q; actual reply %q", ping, buf[:n])
	}

	// To make sure, we set an ample deadline.
	err = client.SetDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	// And attempt to read another message.
	_, err = client.Read(buf)
	if err == nil {
		t.Fatal("unexpected packet")
	}
}