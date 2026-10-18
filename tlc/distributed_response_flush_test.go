package tlc

import (
	"io"
	"net"
	"net/rpc"
	"sync/atomic"
	"testing"
	"time"
)

type distributedFailedResponseConn struct {
	net.Conn
	writes atomic.Int32
}

func (c *distributedFailedResponseConn) Write([]byte) (int, error) {
	c.writes.Add(1)
	return 0, io.ErrClosedPipe
}

// Native Go codec ownership has no corresponding original Java test method.
// A failed response must unblock the peer and the serving goroutine, even when
// the connection's read direction remains open. The operation is not replayed.
func TestDistributedResponseFlushFailureClosesTransport(t *testing.T) {
	host := NewDistributedRPCServer()
	endpoint := &distributedPartialRequestEndpoint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(NewMemFPSet())}
	if err := host.RegisterFingerprint("primary", endpoint); err != nil {
		t.Fatal(err)
	}
	serverConn, peer := net.Pipe()
	conn := &distributedFailedResponseConn{Conn: serverConn}
	client := rpc.NewClient(peer)
	served := make(chan struct{})
	go func() {
		host.rpc.ServeCodec(newDistributedServerCodec(host, conn))
		close(served)
	}()
	defer func() {
		_ = client.Close()
		_ = conn.Close()
		<-served
	}()
	call := client.Go("Fingerprint.Call", DistributedFingerprintRequest{Object: "primary", Operation: "size"}, &DistributedFingerprintReply{}, make(chan *rpc.Call, 1))
	select {
	case completed := <-call.Done:
		if completed.Error == nil {
			t.Fatal("failed response was reported as successful")
		}
	case <-time.After(time.Second):
		t.Fatal("failed response left peer waiting on an open transport")
	}
	select {
	case <-served:
	case <-time.After(time.Second):
		t.Fatal("failed response left server reading the connection")
	}
	host.replies.Wait()
	if endpoint.calls.Load() != 1 || conn.writes.Load() != 1 {
		t.Fatalf("accepted operation or response replayed: calls=%d writes=%d", endpoint.calls.Load(), conn.writes.Load())
	}
}
