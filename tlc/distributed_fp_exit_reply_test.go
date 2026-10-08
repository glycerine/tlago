package tlc

import (
	"errors"
	"io"
	"net/rpc"
	"os"
	"sync/atomic"
	"testing"
)

type closingFingerprintExit struct {
	*LocalFingerprintEndpoint
	host  atomic.Pointer[DistributedRPCServer]
	calls atomic.Int64
}

func (e *closingFingerprintExit) Exit(cleanup bool) error {
	e.calls.Add(1)
	if err := e.LocalFingerprintEndpoint.Exit(cleanup); err != nil {
		return err
	}
	if host := e.host.Load(); host != nil {
		return host.Close()
	}
	return nil
}

// No original test covers loss of a native exit reply. FPSetManager's source
// close loop tolerates the server disappearing during exit and visits later
// registrations. Other fingerprint operations retain ordinary failure handling.
func TestFingerprintRPCCompletedExitLostReply(t *testing.T) {
	for _, throughManager := range []bool{false, true} {
		name := "direct"
		if throughManager {
			name = "manager"
		}
		t.Run(name, func(t *testing.T) {
			storage := NewMemFPSet()
			storage.Init(1, t.TempDir(), "fingerprints")
			endpoint := &closingFingerprintExit{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(storage)}
			host, client := startFingerprintRPC(t, endpoint)
			t.Cleanup(func() { _ = host.CloseGracefully() })
			endpoint.host.Store(host)
			if throughManager {
				later := &closingFingerprintExit{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(NewMemFPSet())}
				manager := NewDistributedFPSetManager(client, later)
				output, err := os.CreateTemp(t.TempDir(), "exit-stderr-")
				if err != nil {
					t.Fatal(err)
				}
				previous := os.Stderr
				os.Stderr = output
				t.Cleanup(func() { os.Stderr = previous; _ = output.Close() })
				if err := manager.Close(false); err != nil {
					t.Fatal(err)
				}
				os.Stderr = previous
				if later.calls.Load() != 1 {
					t.Fatal("lost exit reply skipped later registration")
				}
				data, err := os.ReadFile(output.Name())
				if err != nil || len(data) != 0 {
					t.Fatalf("expected exit closure was reported: %v/%s", err, data)
				}
			} else {
				err := client.Exit(false)
				failure, ok := err.(*DistributedOperationError)
				// net/rpc translates EOF for a pending, unclosed call into
				// ErrUnexpectedEOF; retain that exact native cause.
				if !ok || !errors.Is(err, io.ErrUnexpectedEOF) || !failure.Remote || !failure.IO || !failure.ExitIgnorable || failure.Recoverable || failure.WorkerUnavailable {
					t.Fatalf("lost exit reply traits/cause: %T/%v", err, err)
				}
			}
			if err := host.CloseGracefully(); err != nil {
				t.Fatal(err)
			}
			if endpoint.calls.Load() != 1 {
				t.Fatal("exit was replayed")
			}
			err := client.Exit(false)
			failure, ok := err.(*DistributedOperationError)
			if !ok || !errors.Is(err, rpc.ErrShutdown) || failure.ExitIgnorable || endpoint.calls.Load() != 1 {
				t.Fatalf("prior closed connection was ignored or replayed: %T/%v", err, err)
			}
		})
	}
}
