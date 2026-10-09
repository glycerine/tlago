package tlc

import (
	"errors"
	"net/rpc"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Native coordinator views share connection ownership. Closing another view
// or the discovery cache must join primary release before callbacks, retaining
// both failures for later process shutdown. No upstream Java test covers this.
func TestDistributedCoordinatorViewsJoinConnectionRelease(t *testing.T) {
	primaryFailure := errors.New("coordinator transport release failed")
	callbackFailure := errors.New("coordinator callback release failed")
	codec := &distributedGatedClientCloseCodec{
		failedDistributedClientCodec: &failedDistributedClientCodec{requests: make(chan uint64), closed: make(chan struct{})},
		entered:                      make(chan struct{}), release: make(chan struct{}), readDone: make(chan struct{}), failure: primaryFailure,
	}
	client := rpc.NewClientWithCodec(codec)
	endpoint := newNetworkServerEndpoint(client, "coordinator", "original")
	// Same connection/owner, distinct captured endpoint identity, as discovery
	// constructs after a binding is replaced on its existing TCP listener.
	view := &NetworkServerEndpoint{client: client, Address: endpoint.Address, Object: "replacement", children: endpoint.children}
	var callbackCloses atomic.Int32
	if err := endpoint.children.add(distributedConnectionCloser(func() error {
		callbackCloses.Add(1)
		return callbackFailure
	})); err != nil {
		t.Fatal(err)
	}
	var releaseOnce sync.Once
	var done []<-chan struct{}
	t.Cleanup(func() {
		releaseOnce.Do(func() { close(codec.release) })
		for _, joined := range append(done, codec.readDone) {
			select {
			case <-joined:
			case <-time.After(2 * time.Second):
				t.Error("coordinator cleanup did not join after release")
			}
		}
	})
	start := func(releaseCall func() error) <-chan error {
		result, joined := make(chan error, 1), make(chan struct{})
		done = append(done, joined)
		go func() { defer close(joined); result <- releaseCall() }()
		return result
	}
	first := start(endpoint.CloseConnection)
	select {
	case <-codec.entered:
	case <-time.After(2 * time.Second):
		t.Fatal("coordinator never began primary release")
	}
	second := start(view.CloseConnection)
	select {
	case err := <-second:
		t.Fatalf("other view returned before primary release: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	if callbackCloses.Load() != 0 {
		t.Fatal("callback cleanup ran before primary release completed")
	}
	releaseOnce.Do(func() { close(codec.release) })
	for _, result := range []<-chan error{first, second} {
		select {
		case err := <-result:
			if !errors.Is(err, primaryFailure) || !errors.Is(err, callbackFailure) {
				t.Errorf("view close lost primary/callback result: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("view close did not finish after release")
		}
	}
	discovery := NewDistributedNetworkDiscovery()
	discovery.clients["coordinator"] = endpoint
	if err := discovery.Close(); !errors.Is(err, primaryFailure) || !errors.Is(err, callbackFailure) {
		t.Fatalf("later discovery close lost earlier failures: %v", err)
	}
	if codec.closes.Load() != 1 || callbackCloses.Load() != 1 || codec.writes.Load() != 0 {
		t.Fatalf("release repeated or issued RPC: primary=%d callbacks=%d writes=%d", codec.closes.Load(), callbackCloses.Load(), codec.writes.Load())
	}
}
