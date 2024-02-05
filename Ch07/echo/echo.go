// Pages 144 - 145
// Listing 7-1: Creating the streaming echo server function.
package echo

import (
	"context"
	"net"
)

func streamingEchoServer(ctx context.Context, network string, addr string) (net.Addr, error) {
	s, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
}
