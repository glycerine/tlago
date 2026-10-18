package tlc

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type distributedShutdownAcceptListener struct {
	conn    net.Conn
	failure error
	entered chan struct{}
	release chan struct{}
	once    sync.Once
	closes  atomic.Int32
}

func (l *distributedShutdownAcceptListener) Accept() (net.Conn, error) {
	close(l.entered)
	<-l.release
	return l.conn, l.failure
}
func (l *distributedShutdownAcceptListener) Close() error {
	l.closes.Add(1)
	l.once.Do(func() { close(l.release) })
	return nil
}
func (*distributedShutdownAcceptListener) Addr() net.Addr { return nil }

type distributedShutdownAcceptedConn struct {
	net.Conn
	failure error
	closes  atomic.Int32
}

func (c *distributedShutdownAcceptedConn) Close() error {
	c.closes.Add(1)
	return errors.Join(c.Conn.Close(), c.failure)
}

// These native shutdown ownership boundaries have no direct Java test method.
func TestDistributedServeAfterCloseRetainsListenerFailure(t *testing.T) {
	for _, mixed := range []bool{false, true} {
		t.Run(map[bool]string{false: "closed_only", true: "closed_and_failure"}[mixed], func(t *testing.T) {
			host := NewDistributedRPCServer()
			if err := host.Close(); err != nil {
				t.Fatal(err)
			}
			failure := errors.New("rejected listener cleanup failed")
			closeErr := error(net.ErrClosed)
			if mixed {
				closeErr = errors.Join(net.ErrClosed, failure)
			}
			listener := &distributedFailingListener{closeFailure: closeErr}
			err := host.Serve(listener)
			if !errors.Is(err, net.ErrClosed) || errors.Is(err, failure) != mixed || listener.closes.Load() != 1 {
				t.Fatalf("rejected listener lost shutdown causes or ownership: %v, closes %d", err, listener.closes.Load())
			}
		})
	}
}

func TestDistributedServeShutdownRetainsLateAcceptFailure(t *testing.T) {
	for _, boundary := range []string{"accept", "connection_close"} {
		for _, mixed := range []bool{false, true} {
			t.Run(boundary+"/"+map[bool]string{false: "closed_only", true: "closed_and_failure"}[mixed], func(t *testing.T) {
				host := NewDistributedRPCServer()
				failure := errors.New("late accept cleanup failed")
				closeErr := error(net.ErrClosed)
				if mixed {
					closeErr = errors.Join(net.ErrClosed, failure)
				}
				listener := &distributedShutdownAcceptListener{entered: make(chan struct{}), release: make(chan struct{})}
				var accepted *distributedShutdownAcceptedConn
				if boundary == "accept" {
					listener.failure = closeErr
				} else {
					conn, peer := net.Pipe()
					t.Cleanup(func() { _ = conn.Close(); _ = peer.Close() })
					accepted = &distributedShutdownAcceptedConn{Conn: conn, failure: closeErr}
					listener.conn = accepted
				}
				done, joined := make(chan error, 1), make(chan struct{})
				go func() { defer close(joined); done <- host.Serve(listener) }()
				t.Cleanup(func() { _ = host.Close(); <-joined })
				select {
				case <-listener.entered:
				case <-time.After(time.Second):
					t.Fatal("Serve did not enter Accept")
				}
				if err := host.Close(); err != nil {
					t.Fatal(err)
				}
				select {
				case err := <-done:
					if mixed && !errors.Is(err, failure) || !mixed && err != nil {
						t.Fatalf("late %s lost shutdown failure or reported closed-only error: %v", boundary, err)
					}
				case <-time.After(time.Second):
					t.Fatal("Serve did not finish after shutdown")
				}
				if listener.closes.Load() != 1 || accepted != nil && accepted.closes.Load() != 1 {
					t.Fatal("shutdown repeated or skipped listener/late connection release")
				}
				host.mu.Lock()
				listeners, connections := len(host.listeners), len(host.connections)
				host.mu.Unlock()
				if listeners != 0 || connections != 0 {
					t.Fatal("shutdown retained listener or admitted a late connection")
				}
			})
		}
	}
}
