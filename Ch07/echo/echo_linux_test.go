// Page 152
// Listing 7-10: Instantiating a packet-based streaming echo server.
// *Note: Windows, WSL, and macOS do not support unixpacket domain sockets.
// Notice that this file's suffix "_linux_test.go" is a build constraint that
// informs Go that it should only invoke this test when running on Linux.
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

func TestEchoServerUnixPacket(t *testing.T) {
	dir, err := os.MkdirTemp("", "echo_unixpacket")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	socket := filepath.Join(dir, fmt.Sprintf("%d.sock", os.Getpid()))
	rAddr, err := streamingEchoServer(ctx, "unixpacket", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	err = os.Chmod(socket, os.ModeSocket|0666)
	if err != nil {
		t.Fatal(err)
	}

	// Pages 152-153
	// Listing 7-11: Using a unixpacket socket to echo messages
	// Since unixpacket is session oriented, we use net.Dial to initiate a
	// connection with the server.
	// We do not simply write to the server's address, as we would if the
	// network type were datagram based.
	conn, err := net.Dial("unixpacket", rAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	msg := []byte("ping")
	// We can see the distinction between the unix and unixpacket socket types
	// by writing three ping messages to the server before reading the first reply.
	for i := 0; i < 3; i++ { // write 3 "ping" messages
		_, err = conn.Write(msg)
		if err != nil {
			t.Fatal(err)
		}
	}

	buf := make([]byte, 1024)
	// Whereas a unix socket type would return all three ping messages with a
	// single read, unixpacket acts just like other datagram-based network types
	// and returns one message for each read operation.
	for i := 0; i < 3; i++ { // read 3 times from the server
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(msg, buf[:n]) {
			t.Errorf("expected reply %q; actual reply %q", msg, buf[:n])
		}
	}

	// Pages 153-154
	// Listing 7-12:
	// This illustrates how unixpacket discards unrequested data in each
	// datagram.
	for i := 0; i < 3; i++ { // write 3 more "ping" messages
		_, err = conn.Write(msg)
		if err != nil {
			t.Fatal(err)
		}
	}

	// This time around, we reduce the buffer size to only 2 bytes and read the
	// first 2 bytes of each datagram.
	buf = make([]byte, 2)    // read only the first 2 bytes of each reply
	for i := 0; i < 3; i++ { // read 3 times from the server
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		// If we were using a streaming network type like tcp or unix, we would
		// expect to read pi for the first read and ng for the second read.
		// But unixpacket discards the ng portion of the ping message because we
		// requested only the first two bytes (pi).
		// Therefore, we make sure we're only receiving the first 2 bytes of the
		// datagram with each read.
		if !bytes.Equal(msg[:2], buf[:n]) {
			t.Errorf("expected reply %q; actual reply %q", msg[:2], buf[:n])
		}
	}
}
