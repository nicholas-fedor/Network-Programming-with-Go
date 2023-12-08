// Page 59
// Listing 3-7: Directly canceling the context to abort the connection attempt.
package ch03

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContextCancel(t *testing.T) {
	// Instead of creating a context with a deadline and waiting for the
	// deadline to abort the connection attempt, you use context.WithCancel to
	// return a context and a function to cancel the context.
	ctx, cancel := context.WithCancel(context.Background())
	sync := make(chan struct{})

	// Since you're manually canceling the context, you create a closure and
	// spin it off in a goroutine to handle the connection attempt.
	go func() {
		defer func() {
			sync <- struct{}{}
		}()
		var d net.Dialer
		d.Control = func(_, _ string, _ syscall.RawConn) error {
			time.Sleep(time.Second)
			return nil
		}
		conn, err := d.DialContext(ctx, "tcp", "10.0.0.1:80")
		if err != nil {
			t.Log(err)
			return
		}

		conn.Close()
		t.Error("connection did not time out")
	}()

	// Once the dialer is attempting to connect to and handshake with the remote
	// node, you call the cancel function to cancel the context.
	// This causes the DialContext method to immediately return with a non-nil
	// error, exiting the goroutine.
	cancel()
	<-sync

	// You can check the context's Err method to make sure the call to cancel
	// was what resulted in the canceled context, as opposed to a deadline in
	// Listing 3-6.
	// In this case, the context's Err method should return a context.Canceled error.
	if ctx.Err() != context.Canceled {
		t.Errorf("expected cancelled context; actual: %q", ctx.Err())
	}
}
