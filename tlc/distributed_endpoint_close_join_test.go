package tlc

import (
	"errors"
	"net/rpc"
	"sync"
	"testing"
	"time"
)

type distributedGatedClientCloseCodec struct {
	*failedDistributedClientCodec
	entered, release, readDone chan struct{}
	failure                    error
}

func (c *distributedGatedClientCloseCodec) Close() error {
	close(c.entered)
	<-c.release
	_ = c.failedDistributedClientCodec.Close()
	return c.failure
}

func (c *distributedGatedClientCloseCodec) ReadResponseHeader(response *rpc.Response) error {
	defer close(c.readDone)
	return c.failedDistributedClientCodec.ReadResponseHeader(response)
}

// Native client release has no original Java test. A direct endpoint close
// racing its callback owner's close must join the same release and preserve
// its failure, while calls reject the closed endpoint without dialing/replay.
func TestDistributedEndpointCloseJoinsCallbackOwner(t *testing.T) {
	for _, role := range []string{"worker", "fingerprint"} {
		t.Run(role, func(t *testing.T) {
			failure := errors.New("endpoint transport release failed")
			codec := &distributedGatedClientCloseCodec{
				failedDistributedClientCodec: &failedDistributedClientCodec{requests: make(chan uint64), closed: make(chan struct{})},
				entered:                      make(chan struct{}), release: make(chan struct{}), readDone: make(chan struct{}), failure: errors.Join(rpc.ErrShutdown, failure),
			}
			client := rpc.NewClientWithCodec(codec)
			var closeEndpoint func() error
			var acquire func() (*rpc.Client, error)
			if role == "worker" {
				endpoint := &NetworkWorkerEndpoint{client: client}
				closeEndpoint, acquire = endpoint.CloseConnection, endpoint.clientForCall
			} else {
				endpoint := &NetworkFingerprintEndpoint{client: client}
				closeEndpoint, acquire = endpoint.CloseConnection, endpoint.clientForCall
			}
			var owner distributedConnections
			if err := owner.add(distributedConnectionCloser(closeEndpoint)); err != nil {
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
						t.Error("endpoint cleanup did not join after release")
					}
				}
			})
			start := func(releaseCall func() error) <-chan error {
				result, joined := make(chan error, 1), make(chan struct{})
				done = append(done, joined)
				go func() { defer close(joined); result <- releaseCall() }()
				return result
			}
			first := start(closeEndpoint)
			select {
			case <-codec.entered:
			case <-time.After(2 * time.Second):
				t.Fatal("endpoint never began transport release")
			}
			if acquired, err := acquire(); acquired != nil || err != rpc.ErrShutdown || codec.writes.Load() != 0 {
				t.Fatalf("closed endpoint admitted a call: %v/%v", acquired, err)
			}
			second := start(owner.close)
			select {
			case err := <-second:
				t.Fatalf("callback owner returned before endpoint release: %v", err)
			case <-time.After(50 * time.Millisecond):
			}
			releaseOnce.Do(func() { close(codec.release) })
			for _, result := range []<-chan error{first, second} {
				select {
				case err := <-result:
					if !errors.Is(err, failure) {
						t.Errorf("endpoint/owner lost original close failure: %v", err)
					}
				case <-time.After(2 * time.Second):
					t.Fatal("endpoint/owner did not join release")
				}
			}
			if err := closeEndpoint(); !errors.Is(err, failure) || codec.closes.Load() != 1 {
				t.Fatalf("later endpoint close lost failure or repeated release: %v, closes=%d", err, codec.closes.Load())
			}
		})
	}
}
