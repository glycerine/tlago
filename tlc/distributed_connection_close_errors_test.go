package tlc

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"testing"
	"time"
)

// Native ownership has no upstream Java test. Already-closed components must
// not hide a different connection's failure inside an errors.Join result.
func TestDistributedConnectionCloseRetainsMixedFailures(t *testing.T) {
	failure := errors.New("owned callback close failed")
	for _, closeErr := range []error{
		failure,
		errors.Join(net.ErrClosed, failure),
		errors.Join(rpc.ErrShutdown, failure),
		fmt.Errorf("callback cleanup: %w", errors.Join(net.ErrClosed, rpc.ErrShutdown, failure)),
	} {
		t.Run(closeErr.Error(), func(t *testing.T) {
			var connections distributedConnections
			closed := 0
			if err := connections.add(distributedConnectionCloser(func() error { closed++; return closeErr })); err != nil {
				t.Fatal(err)
			}
			if err := connections.add(distributedConnectionCloser(func() error { closed++; return nil })); err != nil {
				t.Fatal(err)
			}
			if err := connections.close(); !errors.Is(err, failure) || closed != 2 {
				t.Fatalf("cleanup lost failure or skipped later owner: %v, closed %d", err, closed)
			}
			if err := connections.close(); !errors.Is(err, failure) || closed != 2 {
				t.Fatal("repeated owner cleanup lost its original failure or repeated connection closes")
			}
		})
	}
}

func TestDistributedConnectionCloseIgnoresOnlyClosedFailures(t *testing.T) {
	for _, closeErr := range []error{nil, net.ErrClosed, rpc.ErrShutdown,
		fmt.Errorf("callback cleanup: %w", net.ErrClosed),
		errors.Join(net.ErrClosed, rpc.ErrShutdown),
	} {
		var connections distributedConnections
		if err := connections.add(distributedConnectionCloser(func() error { return closeErr })); err != nil {
			t.Fatal(err)
		}
		if err := connections.close(); err != nil {
			t.Fatalf("already-closed cleanup should succeed: %v", err)
		}
	}
}

func TestDistributedDiscoveryCloseRetainsCallbackFailure(t *testing.T) {
	fixture := coordinatorFixture()
	_, client := startCoordinatorRPC(t, fixture)
	failure := errors.New("fingerprint callback close failed")
	if err := client.children.add(distributedConnectionCloser(func() error { return failure })); err != nil {
		t.Fatal(err)
	}
	// Close the primary connection first, leaving its owned callbacks for the
	// discovery owner. CloseConnection now joins ErrShutdown with their failure.
	if err := client.client.Close(); err != nil {
		t.Fatal(err)
	}
	discovery := NewDistributedNetworkDiscovery()
	discovery.clients["coordinator"] = client
	if err := discovery.Close(); !errors.Is(err, failure) {
		t.Fatalf("discovery discarded callback failure beside primary shutdown: %v", err)
	}
	if err := discovery.Close(); err != nil {
		t.Fatal(err)
	}
}

// The native host owns listener shutdown, read interruption and connection
// cleanup. An already-closed cause must not hide a different failure at any of
// those boundaries. There is no corresponding upstream Java transport test.
type distributedCloseFailureConn struct {
	net.Conn
	closeFailure    error
	deadlineFailure error
}

func (c *distributedCloseFailureConn) Close() error {
	return errors.Join(c.Conn.Close(), c.closeFailure)
}

func (c *distributedCloseFailureConn) SetReadDeadline(deadline time.Time) error {
	return errors.Join(c.Conn.SetReadDeadline(deadline), c.deadlineFailure)
}

func TestDistributedRPCHostCloseRetainsMixedFailures(t *testing.T) {
	for _, boundary := range []string{"listener", "read_deadline", "connection"} {
		for _, mixed := range []bool{false, true} {
			name := boundary + "/closed_only"
			if mixed {
				name = boundary + "/closed_and_failure"
			}
			t.Run(name, func(t *testing.T) {
				host := NewDistributedRPCServer()
				failure := errors.New("native host cleanup failed")
				closeErr := error(net.ErrClosed)
				if mixed {
					closeErr = fmt.Errorf("native shutdown: %w", errors.Join(net.ErrClosed, failure))
				}
				if boundary == "listener" {
					listener := &distributedFailingListener{closeFailure: closeErr}
					host.listeners[listener] = struct{}{}
				} else {
					conn, peer := net.Pipe()
					t.Cleanup(func() { _ = conn.Close(); _ = peer.Close() })
					owned := &distributedCloseFailureConn{Conn: conn}
					if boundary == "connection" {
						owned.closeFailure = closeErr
					} else {
						owned.deadlineFailure = closeErr
					}
					host.connections[owned] = struct{}{}
				}
				err := host.CloseGracefully()
				if mixed {
					if !errors.Is(err, failure) || !errors.Is(err, net.ErrClosed) {
						t.Fatalf("host lost mixed %s cleanup causes: %v", boundary, err)
					}
				} else if err != nil {
					t.Fatalf("host reported benign %s shutdown: %v", boundary, err)
				}
				// Connection cleanup is cached by the host and must not lose
				// its failure on a later close either.
				if boundary == "connection" {
					err = host.Close()
					if mixed && !errors.Is(err, failure) {
						t.Fatalf("repeated host close lost connection failure: %v", err)
					}
					if !mixed && err != nil {
						t.Fatalf("repeated host close reported benign failure: %v", err)
					}
				}
			})
		}
	}
}
