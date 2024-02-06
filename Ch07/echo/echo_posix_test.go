//go:build darwin || linux
// +build darwin linux

// Page 149
// Listing 7-6: Building constraints and imports for macOS and Linux
package echo

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
)

// Pages 149-150
// Listing 7-7: Instantiating the datagram-based echo server.
func TestEchoServerUnixDatagram(t *testing.T) {
	dir, err := os.MkdirTemp("", "echo_unixgram")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	// Just as with UDP connections, both the server and the client must bind to
	// an address so they can send and receive datagrams.
	// The server has its own socket file that is separate from the client's
	// socket file in Listing 7-8.
	sSocket := filepath.Join(dir, fmt.Sprintf("s%d.sock", os.Getpid()))
	serverAddr, err := datagramEchoServer(ctx, "unixgram", sSocket)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	err = os.Chmod(sSocket, os.ModeSocket|0622)
	if err != nil {
		t.Fatal(err)
	}

	// Page 150
	// Listing 7-8: Instantiating the datagram-based echo client.
	// Just as with UDP connections, both the server and the client must bind to
	// an address so they can send and receive datagrams.
	// The server has its own socket file that is separate from the client's
	// socket file in Listing 7-8.
	cSocket := filepath.Join(dir, fmt.Sprintf("c%d.sock", os.Getpid()))
	client, err := net.ListenPacket("unixgram", cSocket)
	if err != nil {
		t.Fatal(err)
	}

	// The call to os.Remove in Listing 7-5's datagramEchoServer function cleans
	// up the socket file when the server closes.
	// The client has some additional housecleaning, so we make the client clean
	// up its own socket file when it's all done listening to it.
	defer func() { _ = client.Close() }()

	// The server should be able to write to the client's socket file as well as
	// as well as its own socket file, or the server won't be able to reply to
	// any message.
	err = os.Chmod(cSocket, os.ModeSocket|0622)
	if err != nil {
		t.Fatal(err)
	}

	// Page 151
	// Listing 7-9: Using unixgram sockets to echo messages.
	msg := []byte("ping")
	for i := 0; i < 3; i++ { // Write 3 "ping" messages
		// We write three ping messages to the server before reading the first datagram.
		_, err = client.WriteTo(msg, serverAddr)
		if err != nil {
			t.Fatal(err)
		}
	}

	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ { // read 3 "ping" messages
		// We then perform three reads with a buffer large enough to hold all
		// three ping messages. As expected, unixgram sockets maintain the
		// delineation between messages; we send three messages and read three replies.
		n, addr, err := client.ReadFrom(buf)
		if err != nil {
			t.Fatal(err)
		}

		if addr.String() != serverAddr.String() {
			t.Fatalf("received reply from %q instead of %q", addr, serverAddr)
		}

		if !bytes.Equal(msg, buf[:n]) {
			t.Fatalf("expected reply %q; actual reply %q", msg, buf[:n])
		}
	}
}
