package tlc

import (
	"errors"
	"net"
	"testing"
	"time"
)

// Accepted native connections need one close owner even when net/rpc, its
// serving goroutine and host shutdown all release them. No Java test exists
// for these Go transport ownership boundaries.
func TestDistributedAcceptedConnectionCloseOwnership(t *testing.T) {
	for _, initiator := range []string{"peer", "host"} {
		for _, mixed := range []bool{false, true} {
			t.Run(initiator+"/"+map[bool]string{false: "closed_only", true: "closed_and_failure"}[mixed], func(t *testing.T) {
				host := NewDistributedRPCServer()
				conn, peer := net.Pipe()
				t.Cleanup(func() { _ = peer.Close(); _ = host.Close(); _ = conn.Close() })
				failure := errors.New("accepted connection release failed")
				closeErr := error(net.ErrClosed)
				if mixed {
					closeErr = errors.Join(net.ErrClosed, failure)
				}
				accepted := &distributedShutdownAcceptedConn{Conn: conn, failure: closeErr}
				acceptFailure := errors.New("stop listener after accepting connection")
				listener := &distributedFailingListener{conn: accepted, failure: acceptFailure}
				if err := host.Serve(listener); !errors.Is(err, acceptFailure) {
					t.Fatalf("listener did not accept and stop: %v", err)
				}
				if initiator == "peer" {
					if err := peer.Close(); err != nil {
						t.Fatal(err)
					}
				} else {
					err := host.Close()
					if mixed && !errors.Is(err, failure) || !mixed && err != nil {
						t.Fatalf("host-initiated close lost failure or reported benign closure: %v", err)
					}
				}
				deadline := time.Now().Add(time.Second)
				for {
					host.mu.Lock()
					remaining := len(host.connections)
					host.mu.Unlock()
					if remaining == 0 {
						break
					}
					if time.Now().After(deadline) {
						t.Fatal("accepted connection cleanup did not finish")
					}
					time.Sleep(time.Millisecond)
				}
				for _, closeHost := range []func() error{host.Close, host.CloseGracefully} {
					err := closeHost()
					if mixed && (!errors.Is(err, failure) || !errors.Is(err, net.ErrClosed)) || !mixed && err != nil {
						t.Fatalf("host forgot removed connection failure or reported benign closure: %v", err)
					}
				}
				if accepted.closes.Load() != 1 {
					t.Fatalf("accepted connection closed %d times", accepted.closes.Load())
				}
			})
		}
	}
}
