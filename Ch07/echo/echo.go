// Pages 144 - 145
// Listing 7-1: Creating the streaming echo server function.
package echo

import (
	"context"
	"net"
)

// A listener created with either net.Listen or net.ListenUnix will
// automatically remove the socket file when the listener exits. You can modify
// this behavior with net.UnixListener's SetUnlinkOnClose method, though the
// default is ideal for most cases. Unix domain socket files created with
// net.ListenPacket won't be automatically removed when the listener exits, as
// we'll see later.
// As before, we spin off the echo server in its own goroutine so it can
// asynchronously accept connections.
func streamingEchoServer(ctx context.Context, network string, addr string) (net.Addr, error) {
	s, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}

	// Pages 145-146
	// Listing 7-2: A stream-based echo server.
	go func() {
		go func() {
			// Once the caller cancels the context, the server closes.
			<-ctx.Done()
			_ = s.Close()
		}()

		for {
			// Once the server accepts a connection, we start a goroutine to
			// echo incoming messages.
			conn, err := s.Accept()
			if err != nil {
				return
			}
			
			go func() {
				defer func() { _ = conn.Close() }()
				
				// Since we're using net.Conn interface, we can use its Read and
				// Write methods to communicate with the client no matter whether
				// the server is communicating over a network socket or a Unix
				// domain socket.
				for {
					buf := make([]byte, 1024)
					n, err := conn.Read(buf)
					if err != nil {
						return
					}

					_, err = conn.Write(buf[:n])
					if err != nil {
						return
					}
				}
			}()
		}
	}()

	return s.Addr(), nil
}
