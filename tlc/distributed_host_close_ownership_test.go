package tlc

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type distributedOwnedCloseListener struct {
	close func() error
}

func (*distributedOwnedCloseListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (*distributedOwnedCloseListener) Addr() net.Addr            { return nil }
func (l *distributedOwnedCloseListener) Close() error            { return l.close() }

type distributedReadInterruptionFailure struct {
	net.Conn
	failure error
	calls   atomic.Int32
	entered chan struct{}
	release chan struct{}
}

func (c *distributedReadInterruptionFailure) SetReadDeadline(deadline time.Time) error {
	err := c.Conn.SetReadDeadline(deadline)
	if c.calls.Add(1) == 1 {
		if c.entered != nil {
			close(c.entered)
			<-c.release
		}
		return errors.Join(err, c.failure)
	}
	return err
}

// These checks cover native transport ownership; no original Java test exists.
func TestDistributedHostCloseRetainsEarlierShutdownFailures(t *testing.T) {
	for _, boundary := range []string{"listener", "read_interruption"} {
		t.Run(boundary, func(t *testing.T) {
			host := NewDistributedRPCServer()
			failure := errors.New("first shutdown failed")
			if boundary == "listener" {
				listener := &distributedOwnedCloseListener{close: func() error { return failure }}
				host.listeners[listener] = struct{}{}
			} else {
				conn, peer := net.Pipe()
				t.Cleanup(func() { _ = conn.Close(); _ = peer.Close() })
				host.connections[&distributedReadInterruptionFailure{Conn: conn, failure: failure}] = struct{}{}
			}
			if err := host.CloseGracefully(); !errors.Is(err, failure) {
				t.Fatalf("initial close lost shutdown failure: %v", err)
			}
			// A real Serve loop removes a closed listener from host tracking.
			host.mu.Lock()
			clear(host.listeners)
			host.mu.Unlock()
			for _, closeHost := range []func() error{host.Close, host.CloseGracefully} {
				if err := closeHost(); !errors.Is(err, failure) {
					t.Fatalf("later close lost earlier %s failure: %v", boundary, err)
				}
			}
		})
	}
}

func TestDistributedHostCloseJoinsListenerShutdown(t *testing.T) {
	for _, graceful := range []bool{false, true} {
		t.Run(map[bool]string{false: "forced", true: "graceful"}[graceful], func(t *testing.T) {
			host := NewDistributedRPCServer()
			entered, release := make(chan struct{}), make(chan struct{})
			var releaseOnce sync.Once
			var closing sync.WaitGroup
			var calls atomic.Int32
			failure := errors.New("listener release failed")
			listener := &distributedOwnedCloseListener{close: func() error {
				if calls.Add(1) == 1 {
					close(entered)
					<-release
					return failure
				}
				return net.ErrClosed
			}}
			host.listeners[listener] = struct{}{}
			first, second := make(chan error, 1), make(chan error, 1)
			t.Cleanup(func() {
				releaseOnce.Do(func() { close(release) })
				closing.Wait()
			})
			closing.Add(1)
			go func() { defer closing.Done(); first <- host.Close() }()
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("host did not start listener shutdown")
			}
			closing.Add(1)
			go func() {
				defer closing.Done()
				if graceful {
					second <- host.CloseGracefully()
				} else {
					second <- host.Close()
				}
			}()
			select {
			case err := <-second:
				t.Fatalf("concurrent close skipped listener shutdown: %v", err)
			case <-time.After(50 * time.Millisecond):
			}
			releaseOnce.Do(func() { close(release) })
			for _, result := range []<-chan error{first, second} {
				select {
				case err := <-result:
					if !errors.Is(err, failure) {
						t.Fatalf("close lost listener shutdown failure: %v", err)
					}
				case <-time.After(time.Second):
					t.Fatal("close did not join listener shutdown")
				}
			}
			if calls.Load() != 1 {
				t.Fatalf("listener closed %d times", calls.Load())
			}
		})
	}
}

func TestDistributedHostForcedCloseDuringReadInterruption(t *testing.T) {
	host := NewDistributedRPCServer()
	conn, peer := net.Pipe()
	failure := errors.New("read interruption failed")
	owned := &distributedReadInterruptionFailure{
		Conn: conn, failure: failure, entered: make(chan struct{}), release: make(chan struct{}),
	}
	host.connections[owned] = struct{}{}
	var releaseOnce sync.Once
	var closing sync.WaitGroup
	t.Cleanup(func() {
		releaseOnce.Do(func() { close(owned.release) })
		closing.Wait()
		_ = conn.Close()
		_ = peer.Close()
	})
	graceful, forced := make(chan error, 1), make(chan error, 1)
	closing.Add(1)
	go func() { defer closing.Done(); graceful <- host.CloseGracefully() }()
	select {
	case <-owned.entered:
	case <-time.After(time.Second):
		t.Fatal("graceful close did not reach read interruption")
	}
	closing.Add(1)
	go func() { defer closing.Done(); forced <- host.Close() }()
	select {
	case err := <-forced:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("forced close waited for graceful read interruption")
	}
	releaseOnce.Do(func() { close(owned.release) })
	select {
	case err := <-graceful:
		if !errors.Is(err, failure) {
			t.Fatalf("graceful close lost read interruption failure: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("graceful close did not finish")
	}
	if err := host.Close(); !errors.Is(err, failure) || owned.calls.Load() != 1 {
		t.Fatalf("later close lost failure or repeated interruption: %v, calls %d", err, owned.calls.Load())
	}
}
