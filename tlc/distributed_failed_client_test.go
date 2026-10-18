package tlc

import (
	"io"
	"net/rpc"
	"sync"
	"sync/atomic"
	"testing"
)

// Native transport ownership has no corresponding Java method test.
type failedDistributedClientCodec struct {
	fatal    bool
	requests chan uint64
	closed   chan struct{}
	once     sync.Once
	writes   atomic.Int32
	closes   atomic.Int32
}

func (c *failedDistributedClientCodec) WriteRequest(request *rpc.Request, _ any) error {
	c.writes.Add(1)
	select {
	case c.requests <- request.Seq:
		return nil
	case <-c.closed:
		return io.ErrClosedPipe
	}
}
func (c *failedDistributedClientCodec) ReadResponseHeader(response *rpc.Response) error {
	select {
	case sequence := <-c.requests:
		if c.fatal {
			return io.EOF
		}
		response.Seq = sequence
		if sequence == 0 {
			response.Error = "remote method failure"
		} else {
			response.Error = ""
		}
		return nil
	case <-c.closed:
		return io.EOF
	}
}
func (*failedDistributedClientCodec) ReadResponseBody(any) error { return nil }
func (c *failedDistributedClientCodec) Close() error {
	c.closes.Add(1)
	c.once.Do(func() { close(c.closed) })
	return nil
}

func TestDistributedFailedClientsReleaseTransport(t *testing.T) {
	for _, role := range []string{"coordinator", "worker", "fingerprint"} {
		for _, fatal := range []bool{true, false} {
			name := "method_error"
			if fatal {
				name = "connection_error"
			}
			t.Run(role+"/"+name, func(t *testing.T) {
				codec := &failedDistributedClientCodec{fatal: fatal, requests: make(chan uint64, 1), closed: make(chan struct{})}
				client := rpc.NewClientWithCodec(codec)
				t.Cleanup(func() { _ = client.Close() })
				var call func() error
				switch role {
				case "coordinator":
					endpoint := &NetworkServerEndpoint{client: client}
					call = func() error { _, err := endpoint.call(DistributedServerRequest{Operation: "done"}); return err }
				case "worker":
					endpoint := &NetworkWorkerEndpoint{client: client}
					call = func() error { _, err := endpoint.call(DistributedWorkerRequest{Operation: "alive"}); return err }
				default:
					endpoint := &NetworkFingerprintEndpoint{client: client}
					call = func() error { _, err := endpoint.call(DistributedFingerprintRequest{Operation: "size"}); return err }
				}
				err := call()
				if err == nil || !isDistributedRemoteFailure(err) || codec.writes.Load() != 1 {
					t.Fatalf("first failure = %v, writes = %d", err, codec.writes.Load())
				}
				if fatal {
					if codec.closes.Load() != 1 || javaThrowableCause(err) != io.ErrUnexpectedEOF {
						t.Fatal("broken client retained its transport or changed the original cause")
					}
					if err := call(); javaThrowableCause(err) != rpc.ErrShutdown || codec.writes.Load() != 1 || codec.closes.Load() != 1 {
						t.Fatal("broken client redialed, replayed or closed its codec again")
					}
				} else if codec.closes.Load() != 0 || call() != nil || codec.writes.Load() != 2 {
					t.Fatal("remote method error closed a usable transport")
				}
			})
		}
	}
}
