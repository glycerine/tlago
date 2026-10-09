package tlc

import (
	"errors"
	"net"
	"net/rpc"
	"sync/atomic"
	"testing"
	"time"
)

type distributedFailingListener struct {
	conn         net.Conn
	failure      error
	closeFailure error
	closes       atomic.Int32
}

func (l *distributedFailingListener) Accept() (net.Conn, error) {
	if l.conn != nil {
		conn := l.conn
		l.conn = nil
		return conn, nil
	}
	return nil, l.failure
}
func (l *distributedFailingListener) Close() error { l.closes.Add(1); return l.closeFailure }
func (*distributedFailingListener) Addr() net.Addr { return nil }

// Native listener ownership has no direct original Java test method.
func TestDistributedListenerFailureReleasesListener(t *testing.T) {
	for _, closeFails := range []bool{false, true} {
		name := "accept_failure"
		if closeFails {
			name += "_and_close_failure"
		}
		t.Run(name, func(t *testing.T) {
			host := NewDistributedRPCServer()
			if err := host.RegisterFingerprint("primary", NewLocalFingerprintEndpoint(NewMemFPSet())); err != nil {
				t.Fatal(err)
			}
			serverConn, peer := net.Pipe()
			client := rpc.NewClient(peer)
			t.Cleanup(func() { _ = client.Close(); _ = host.Close() })
			failure := errors.New("listener accept failed")
			closeFailure := errors.New("listener close failed")
			listener := &distributedFailingListener{conn: serverConn, failure: failure}
			if closeFails {
				listener.closeFailure = closeFailure
			}
			err := host.Serve(listener)
			if !errors.Is(err, failure) || (closeFails && !errors.Is(err, closeFailure)) || listener.closes.Load() != 1 {
				t.Fatalf("failed listener not released with original causes: %v, closes=%d", err, listener.closes.Load())
			}
			// Accept failure stops this listener, not already accepted computation.
			call := client.Go("Fingerprint.Call", DistributedFingerprintRequest{Object: "primary", Operation: "size"}, &DistributedFingerprintReply{}, make(chan *rpc.Call, 1))
			select {
			case completed := <-call.Done:
				if completed.Error != nil {
					t.Fatal(completed.Error)
				}
			case <-time.After(time.Second):
				t.Fatal("accept failure interrupted an accepted connection")
			}
			if err := host.Close(); err != nil {
				t.Fatal(err)
			}
			if listener.closes.Load() != 1 {
				t.Fatal("host retained the failed listener")
			}
		})
	}
}

func TestDistributedNetworkCloseRetainsListenerCleanupFailure(t *testing.T) {
	for _, role := range []string{"coordinator", "worker"} {
		t.Run(role, func(t *testing.T) {
			failure := errors.New("listener cleanup failed")
			done := make(chan error, 1)
			done <- errors.Join(net.ErrClosed, failure)
			var closeNetwork func() error
			if role == "coordinator" {
				network := &DistributedCoordinatorNetwork{Host: NewDistributedRPCServer(), done: done}
				closeNetwork = network.Close
			} else {
				network := &DistributedWorkerNetwork{Host: NewDistributedRPCServer(), Discovery: NewDistributedNetworkDiscovery(), done: done}
				closeNetwork = network.Close
			}
			for i := 0; i < 2; i++ {
				if err := closeNetwork(); !errors.Is(err, failure) {
					t.Fatalf("network lost listener cleanup failure: %v", err)
				}
			}
		})
	}
}
