package tlc

import (
	"errors"
	"net"
	"net/rpc"
	"sync/atomic"
	"testing"
)

// Upstream has no direct native transport tests. These checks preserve failure
// causes and forbid replaying fingerprint mutations after connection failure.
func TestFingerprintRPCTransportFailureCauses(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = DialFingerprintEndpoint(address, "primary")
	var network *net.OpError
	if !errors.As(err, &network) {
		t.Fatalf("dial failure discarded its Go network cause: %T/%v", err, err)
	}
	storage := NewMemFPSet()
	_, client := startFingerprintRPC(t, NewLocalFingerprintEndpoint(storage))
	if err := client.CloseConnection(); err != nil {
		t.Fatal(err)
	}
	for _, operation := range []struct {
		name string
		call func() error
	}{
		{"put", func() error { _, err := client.Put(73); return err }},
		{"putBlock", func() error { _, err := client.PutBlock(NewLongVecFrom([]int64{74, 75})); return err }},
		{"containsBlock", func() error { _, err := client.ContainsBlock(NewLongVec()); return err }},
		{"checkpoint", func() error { return client.BeginChkptFile("job") }},
		{"recover", func() error { return client.RecoverFile("job") }},
	} {
		t.Run(operation.name, func(t *testing.T) {
			err := operation.call()
			failure, ok := err.(*DistributedOperationError)
			if !ok || !errors.Is(err, rpc.ErrShutdown) || !failure.Remote || !failure.IO || failure.Recoverable || failure.WorkerUnavailable || failure.ExitIgnorable {
				t.Fatalf("fingerprint connection lost its cause or acquired worker retry traits: %T/%v", err, err)
			}
			if storage.Size() != 0 {
				t.Fatal("failed call replayed a fingerprint mutation")
			}
		})
	}
}

type lostFingerprintReplyEndpoint struct {
	*LocalFingerprintEndpoint
	host  atomic.Pointer[DistributedRPCServer]
	calls atomic.Int64
}

func (e *lostFingerprintReplyEndpoint) PutBlock(fps *LongVec) (*BitVector, error) {
	e.calls.Add(1)
	bits, err := e.LocalFingerprintEndpoint.PutBlock(fps)
	// The insertion has completed, but close the socket before its reply can
	// reach the caller. Repeating it would change the new-fingerprint answer.
	if closeErr := e.host.Load().Close(); closeErr != nil {
		return nil, closeErr
	}
	return bits, err
}

func TestFingerprintRPCCompletedInsertionLostReply(t *testing.T) {
	storage := NewMemFPSet()
	endpoint := &lostFingerprintReplyEndpoint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(storage)}
	host, client := startFingerprintRPC(t, endpoint)
	endpoint.host.Store(host)
	bits, err := client.PutBlock(NewLongVecFrom([]int64{73, 74}))
	failure, ok := err.(*DistributedOperationError)
	if bits != nil || !ok || failure.Cause == nil || !failure.Remote || !failure.IO || failure.Recoverable || failure.ExitIgnorable || endpoint.calls.Load() != 1 || storage.Size() != 2 || !storage.Contains(73) || !storage.Contains(74) {
		t.Fatalf("ambiguous insertion lost storage/cause or was replayed: %v, %T/%v, calls %d", bits, err, err, endpoint.calls.Load())
	}
	if _, err := client.PutBlock(NewLongVecFrom([]int64{73, 74})); !errors.Is(err, rpc.ErrShutdown) || endpoint.calls.Load() != 1 {
		t.Fatalf("failed established connection redialed or replayed: %v, calls %d", err, endpoint.calls.Load())
	}
}
