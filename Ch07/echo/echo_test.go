// Pages 146-147
// Listing 7-3: Setting up an echo server test over a unix domain socket.
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

// Listing 7-3 tests the streaming echo server over a Unix domain socket using
// the unix network type.
func TestEchoServerUnix(t *testing.T) {
	// We create a subdirectory in the operating system's temporary directory
	// named echo_unit that will contain the echo server's socket file.
	dir, err := os.MkdirTemp("", "echo_unix")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		// The deferred call to os.RemoveAll cleans up after the server by
		// removing your temporary subdirectory when the test completes.
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	// We pass a socket file named #.sock, where # is the server's process ID,
	// saved in the temporary subdirectory (/tmp/echo_unix/123.sock) to the
	// streamingEchoServer function.
	socket := filepath.Join(dir, fmt.Sprintf("%d.sock", os.Getpid()))
	rAddr, err := streamingEchoServer(ctx, "unix", socket)
	if err != nil {
		t.Fatal(err)
	}

	// We make sure everyone has read and write access to the socket.
	err = os.Chmod(socket, os.ModeSocket|0666)
	if err != nil {
		t.Fatal(err)
	}

	// Page 147
	// Listing 7-4: Streaming data over a Unix domain socket.
	// We dial the server by using the familiar net.Dial function.
	// It accepts the unix network type and the server's address, which is the
	// full path to the Unix domain socket file.
	conn, err := net.Dial("unix", rAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	msg := []byte("ping")
	// We write three ping messages to the echo server before reading the first response.
	for i := 0; i < 3; i++ { // write 3 "ping" messages
		_, err = conn.Write(msg)
		if err != nil {
			t.Fatal(err)
		}
	}

	buf := make([]byte, 1024)
	// When we read the first response with a buffer large enough to store the
	// three messages we just sent, we receive all three ping messages in a
	// single read as the string pingpingping.
	n, err := conn.Read(buf) // read once from the server
	if err != nil {
		t.Fatal(err)
	}

	// Remember, a stream-based connection does not delineate messages.
	// The onus is on us to determine where one message stops and another one
	// starts when we read a series of bytes from the server.
	expected := bytes.Repeat(msg, 3)
	if !bytes.Equal(expected, buf[:n]) {
		t.Fatalf("expected reply %q; actual reply %q", expected, buf[:n])
	}
}

// Pages 148 - 149
// Listing 7-5: A datagram-based echo server.
// This creates an echo server that will communicate using datagram network
// types, such as UDP and unixgram.
// Whether we're communicating over UDP or a unixgram socket, the server looks
// essentially the same.
// The difference is, we will need to cleanup the socket file with a unixgram listener.
func datagramEchoServer(ctx context.Context, network string, addr string) (net.Addr, error) {
	// We call net.ListenPacket, which returns a net.PacketConn.
	s, err := net.ListenPacket(network, addr)
	if err != nil {
		return nil, err
	}

	go func() {
		go func() {
			<-ctx.Done()
			_ = s.Close()
			if network == "unixgram" {
				// Since we don't use net.Listen or net.ListenUnix to create the
				// listener, Go won't cleanup the socket file for us when the
				// server is finished with it.
				// We must make sure we remove the socket file ourselves, or
				// subsequent attempts to bind to the existing socket file will fail.
				_ = os.Remove(addr)
			}
		}()

		buf := make([]byte, 1024)
		for {
			n, clientAddr, err := s.ReadFrom(buf)
			if err != nil {
				return
			}
			_, err = s.WriteTo(buf[:n], clientAddr)
			if err != nil {
				return
			}
		}
	}()

	return s.LocalAddr(), nil
}