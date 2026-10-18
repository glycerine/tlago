package tlc

import (
	"errors"
	"net/rpc"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Native discovery owns cached coordinator connections. No original Java test
// covers concurrent/repeated closure of this Go transport owner.
func TestDistributedDiscoveryCloseJoinsAndRetainsFailures(t *testing.T) {
	primaryFailure := errors.New("discovery primary release failed")
	callbackFailure := errors.New("discovery callback release failed")
	codec := &distributedGatedClientCloseCodec{
		failedDistributedClientCodec: &failedDistributedClientCodec{requests: make(chan uint64), closed: make(chan struct{})},
		entered:                      make(chan struct{}), release: make(chan struct{}), readDone: make(chan struct{}), failure: errors.Join(rpc.ErrShutdown, primaryFailure),
	}
	client := rpc.NewClientWithCodec(codec)
	endpoint := newNetworkServerEndpoint(client, "coordinator", "original")
	var callbackCloses atomic.Int32
	if err := endpoint.children.add(distributedConnectionCloser(func() error {
		callbackCloses.Add(1)
		return callbackFailure
	})); err != nil {
		t.Fatal(err)
	}
	discovery := NewDistributedNetworkDiscovery()
	discovery.clients["coordinator"] = endpoint
	var releaseOnce sync.Once
	var done []<-chan struct{}
	t.Cleanup(func() {
		releaseOnce.Do(func() { close(codec.release) })
		for _, joined := range append(done, codec.readDone) {
			select {
			case <-joined:
			case <-time.After(2 * time.Second):
				t.Error("discovery cleanup did not join after release")
			}
		}
	})
	start := func() <-chan error {
		result, joined := make(chan error, 1), make(chan struct{})
		done = append(done, joined)
		go func() { defer close(joined); result <- discovery.Close() }()
		return result
	}
	first := start()
	select {
	case <-codec.entered:
	case <-time.After(2 * time.Second):
		t.Fatal("discovery never began primary release")
	}
	second := start()
	select {
	case err := <-second:
		t.Fatalf("concurrent discovery close returned before release: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	if callbackCloses.Load() != 0 {
		t.Fatal("discovery released callbacks before the primary")
	}
	if _, err := discovery.Lookup("tcp://127.0.0.1:1/worker"); !errors.Is(err, rpc.ErrShutdown) {
		t.Fatalf("closing discovery accepted lookup or lost shutdown cause: %v", err)
	}
	releaseOnce.Do(func() { close(codec.release) })
	for _, result := range []<-chan error{first, second} {
		select {
		case err := <-result:
			if !errors.Is(err, primaryFailure) || !errors.Is(err, callbackFailure) {
				t.Errorf("discovery close lost primary/callback failures: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("discovery close did not finish after release")
		}
	}
	if err := discovery.Close(); !errors.Is(err, primaryFailure) || !errors.Is(err, callbackFailure) {
		t.Fatalf("repeated discovery close lost cleanup failures: %v", err)
	}
	if codec.closes.Load() != 1 || callbackCloses.Load() != 1 || codec.writes.Load() != 0 {
		t.Fatalf("discovery repeated release or issued RPC: primary=%d callbacks=%d writes=%d", codec.closes.Load(), callbackCloses.Load(), codec.writes.Load())
	}
}
