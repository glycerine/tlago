package tlc

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Native callback ownership has no original Java test. Endpoint views and
// discovery can close the same owner; all must join and retain its one result.
func TestDistributedCallbackCloseJoinsAndRetainsFailure(t *testing.T) {
	var owner distributedConnections
	entered, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	var calls atomic.Int32
	failure := errors.New("callback release failed")
	if err := owner.add(distributedConnectionCloser(func() error {
		calls.Add(1)
		close(entered)
		<-release
		return failure
	})); err != nil {
		t.Fatal(err)
	}
	var pending []<-chan error
	t.Cleanup(func() {
		releaseOnce.Do(func() { close(release) })
		for _, result := range pending {
			select {
			case <-result:
			case <-time.After(2 * time.Second):
				t.Error("callback cleanup did not join after release")
			}
		}
	})
	start := func() <-chan error {
		result := make(chan error, 1)
		pending = append(pending, result)
		go func() { result <- owner.close() }()
		return result
	}
	first := start()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("callback cleanup never started")
	}
	second := start()
	select {
	case err := <-second:
		pending = pending[:1]
		t.Fatalf("second close returned before callback release: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	releaseOnce.Do(func() { close(release) })
	for _, result := range []<-chan error{first, second} {
		select {
		case err := <-result:
			if !errors.Is(err, failure) {
				t.Errorf("close lost original callback failure: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("close did not join callback release")
		}
	}
	pending = nil
	if err := owner.close(); !errors.Is(err, failure) || calls.Load() != 1 {
		t.Fatalf("later close lost failure or repeated release: %v, calls=%d", err, calls.Load())
	}
}

func TestDistributedCallbackLateAddRetainsReleaseFailure(t *testing.T) {
	for _, mixed := range []bool{false, true} {
		t.Run(map[bool]string{false: "closed_only", true: "mixed"}[mixed], func(t *testing.T) {
			var owner distributedConnections
			if err := owner.close(); err != nil {
				t.Fatal(err)
			}
			failure := errors.New("late callback release failed")
			calls := 0
			err := owner.add(distributedConnectionCloser(func() error {
				calls++
				if mixed {
					return errors.Join(net.ErrClosed, failure)
				}
				return net.ErrClosed
			}))
			if !errors.Is(err, net.ErrClosed) || errors.Is(err, failure) != mixed || calls != 1 {
				t.Fatalf("late admission lost shutdown/release result: %v, calls=%d", err, calls)
			}
		})
	}
}

func TestDistributedDiscoveryCloseRetainsEarlierCallbackFailure(t *testing.T) {
	_, client := startCoordinatorRPC(t, coordinatorFixture())
	failure := errors.New("earlier fingerprint callback release failed")
	if err := client.children.add(distributedConnectionCloser(func() error { return failure })); err != nil {
		t.Fatal(err)
	}
	// Direct endpoint shutdown must not consume the failure before its process
	// discovery owner closes the same endpoint and reports to the CLI.
	if err := client.CloseConnection(); !errors.Is(err, failure) {
		t.Fatalf("direct endpoint close lost callback failure: %v", err)
	}
	discovery := NewDistributedNetworkDiscovery()
	discovery.clients["coordinator"] = client
	if err := discovery.Close(); !errors.Is(err, failure) {
		t.Fatalf("process discovery owner lost earlier callback failure: %v", err)
	}
}
