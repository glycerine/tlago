package tlc

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"testing"
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
			if err := connections.close(); err != nil || closed != 2 {
				t.Fatal("repeated owner cleanup repeated connection closes")
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
