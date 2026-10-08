package tlc

import (
	"errors"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"testing"
)

// The original remote model harness is disabled upstream. These tests exercise
// the new Go TCP boundary directly; they do not replace those original models.
func startFingerprintRPC(t *testing.T, endpoint DistributedFingerprintEndpoint) (*DistributedRPCServer, *NetworkFingerprintEndpoint) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := NewDistributedRPCServer()
	if err := server.RegisterFingerprint("primary", endpoint); err != nil {
		_ = listener.Close()
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Error(err)
		}
		if err := <-done; err != nil && !errors.Is(err, net.ErrClosed) {
			t.Error(err)
		}
	})
	client, err := DialFingerprintEndpoint(listener.Addr().String(), "primary")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.CloseConnection() })
	return server, client
}

type rpcFingerprintRepresentationEndpoint struct {
	*LocalFingerprintEndpoint
	bits *BitVector
}

func (e *rpcFingerprintRepresentationEndpoint) ContainsBlock(*LongVec) (*BitVector, error) {
	return e.bits, nil
}

func TestFingerprintRPCNullAndEmptyWordArrays(t *testing.T) {
	for _, test := range []struct {
		name string
		bits *BitVector
	}{
		{"null_vector", nil},
		{"uninitialized_words", &BitVector{}},
		{"empty_words", NewBitVector(0)},
	} {
		t.Run(test.name, func(t *testing.T) {
			endpoint := &rpcFingerprintRepresentationEndpoint{NewLocalFingerprintEndpoint(NewMemFPSet()), test.bits}
			_, client := startFingerprintRPC(t, endpoint)
			got, err := client.ContainsBlock(NewLongVec())
			if err != nil {
				t.Fatal(err)
			}
			if test.bits == nil {
				if got != nil {
					t.Fatal("null vector became an empty answer")
				}
				return
			}
			if got == nil || got == test.bits || (got.word == nil) != (test.bits.word == nil) {
				t.Fatal("null/empty word-array representation changed across TCP")
			}
			if got.word != nil && NewBitVectorIter(got).Next() != -1 {
				t.Fatal("initialized empty word array is not iterable")
			}
		})
	}
}

func TestFingerprintRPCScalarAndBatchAnswers(t *testing.T) {
	storage := NewMemFPSet()
	_, client := startFingerprintRPC(t, NewLocalFingerprintEndpoint(storage))
	const high = uint64(1) << 63
	if found, err := client.Put(high); err != nil || found {
		t.Fatalf("initial insertion = %v/%v", found, err)
	}
	if found, err := client.Put(high); err != nil || !found {
		t.Fatalf("repeat insertion = %v/%v", found, err)
	}
	fps := NewLongVecFrom([]int64{0, 1, 1, 2, math.MinInt64, -1})
	bits, err := client.PutBlock(fps)
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []bool{true, true, false, true, false, true} {
		if bits.Get(i) != want {
			t.Fatalf("new fingerprint bit %d = %v, want %v", i, bits.Get(i), want)
		}
	}
	// Explicitly issuing another insertion has different answers. Neither
	// client nor server may silently issue that second insertion itself.
	bits, err = client.PutBlock(fps)
	if err != nil || bits.TrueCount() != 0 {
		t.Fatalf("repeat batch bits/error = %v/%v", bits, err)
	}
	bits, err = client.ContainsBlock(fps)
	if err != nil || bits.TrueCount() != 0 {
		t.Fatalf("contains batch bits/error = %v/%v", bits, err)
	}
	if count, err := client.Size(); err != nil || count != 5 || count != storage.Size() {
		t.Fatalf("size = %d/%v, want 5", count, err)
	}
	if found, err := client.Contains(math.MaxUint64); err != nil || !found {
		t.Fatalf("all-bit fingerprint lookup = %v/%v", found, err)
	}
	if valid, err := client.CheckInvariant(); err != nil || !valid {
		t.Fatalf("invariant = %v/%v", valid, err)
	}
	if seen, err := client.GetStatesSeen(); err != nil || seen != storage.GetStatesSeen() {
		t.Fatalf("states seen = %d/%v", seen, err)
	}
	if empty, err := client.ContainsBlock(NewLongVec()); err != nil || empty == nil || empty.TrueCount() != 0 {
		t.Fatalf("empty batch response = %v/%v", empty, err)
	}
	if _, err := client.ContainsBlock(nil); err == nil || isJavaIOException(err) {
		t.Fatalf("null batch was converted to an empty batch or lost its runtime failure: %v", err)
	}
	if count, err := client.Size(); err != nil || count != 5 {
		t.Fatalf("endpoint failed after null batch: %d/%v", count, err)
	}
}

func TestFingerprintRPCFailureAndDisconnect(t *testing.T) {
	storage := NewMemFPSet()
	failed := &rpcFailedFingerprintEndpoint{
		LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(storage),
		failure:                  NewIOException("disk failure"),
	}
	server, client := startFingerprintRPC(t, failed)
	if _, err := client.Put(2); err == nil || err.Error() != "disk failure" || !isJavaIOException(err) {
		t.Fatalf("remote storage error = %v", err)
	}
	if calls := failed.calls.Load(); calls != 1 {
		t.Fatalf("failed insertion calls = %d, want 1", calls)
	}
	// The original FaultyFPSet helper throws an unchecked failure after its
	// first insertion. The server must return that failure and remain alive.
	if err := server.RegisterFingerprint("panicking", NewLocalFingerprintEndpoint(newJavaFaultyFPSet())); err != nil {
		t.Fatal(err)
	}
	panicking, err := DialFingerprintEndpoint(client.Address, "panicking")
	if err != nil {
		t.Fatal(err)
	}
	defer panicking.CloseConnection()
	if _, err := panicking.Put(2); err != nil {
		t.Fatal(err)
	}
	if _, err := panicking.Put(2); err == nil || isJavaIOException(err) {
		t.Fatalf("runtime failure category = %v", err)
	}
	if found, err := panicking.Contains(2); err != nil || !found {
		t.Fatalf("endpoint failed after storage panic: %v/%v", found, err)
	}
	// A transport failure enters the existing manager's failover logic.
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	spare := NewMemFPSet()
	manager := NewDistributedFPSetManager(client, NewLocalFingerprintEndpoint(spare))
	if manager.Put(2) || !spare.Contains(2) || storage.Contains(2) {
		t.Fatal("disconnected endpoint did not fail over to the spare")
	}
	if manager.NumOfAliveServers() != 1 {
		t.Fatal("disconnected endpoint remains available")
	}
}

type rpcFailedFingerprintEndpoint struct {
	*LocalFingerprintEndpoint
	failure error
	calls   atomic.Int64
}

func (e *rpcFailedFingerprintEndpoint) Put(uint64) (bool, error) {
	e.calls.Add(1)
	return false, e.failure
}

func TestFingerprintRPCCheckpointRecovery(t *testing.T) {
	directory := t.TempDir()
	storage := NewMemFPSet()
	storage.Init(1, directory, "primary")
	server, client := startFingerprintRPC(t, NewLocalFingerprintEndpoint(storage))
	if _, err := client.Put(5); err != nil {
		t.Fatal(err)
	}
	if err := client.BeginChkptFile("job"); err != nil {
		t.Fatal(err)
	}
	if err := client.CommitChkptFile("job"); err != nil {
		t.Fatal(err)
	}
	recoveredStorage := NewMemFPSet()
	recoveredStorage.Init(1, directory, "recovered")
	if err := server.RegisterFingerprint("recovered", NewLocalFingerprintEndpoint(recoveredStorage)); err != nil {
		t.Fatal(err)
	}
	recovered, err := DialFingerprintEndpoint(client.Address, "recovered")
	if err != nil {
		t.Fatal(err)
	}
	defer recovered.CloseConnection()
	if err := recovered.RecoverFile("job"); err != nil {
		t.Fatal(err)
	}
	if found, err := recovered.Contains(5); err != nil || !found {
		t.Fatalf("recovered fingerprint = %v/%v", found, err)
	}
	if count, err := recovered.Size(); err != nil || count != 1 {
		t.Fatalf("recovered size = %d/%v, want 1", count, err)
	}
}

func TestFingerprintRPCConcurrentInsertion(t *testing.T) {
	_, client := startFingerprintRPC(t, NewLocalFingerprintEndpoint(NewMemFPSet()))
	const clients, fingerprints = 16, 32
	var done sync.WaitGroup
	var newCount atomic.Int64
	failures := make(chan error, clients)
	for range clients {
		done.Add(1)
		go func() {
			defer done.Done()
			batch := NewLongVec()
			for i := range fingerprints {
				batch.AddElement(int64(i))
			}
			bits, err := client.PutBlock(batch)
			if err != nil {
				failures <- err
				return
			}
			newCount.Add(int64(bits.TrueCount()))
		}()
	}
	done.Wait()
	close(failures)
	for err := range failures {
		t.Error(err)
	}
	if count := newCount.Load(); count != fingerprints {
		t.Fatalf("new fingerprint answers = %d, want %d", count, fingerprints)
	}
	if count, err := client.Size(); err != nil || count != fingerprints {
		t.Fatalf("distinct fingerprints = %d/%v, want %d", count, err, fingerprints)
	}
}

func TestFingerprintRPCObjectIdentityAndLifecycle(t *testing.T) {
	storage := NewMemFPSet()
	server, client := startFingerprintRPC(t, NewLocalFingerprintEndpoint(storage))
	if err := server.RegisterFingerprint("primary", NewLocalFingerprintEndpoint(NewMemFPSet())); err == nil {
		t.Fatal("duplicate registration replaced an existing storage object")
	}
	missing, err := DialFingerprintEndpoint(client.Address, "missing")
	if err != nil {
		t.Fatal(err)
	}
	defer missing.CloseConnection()
	if _, err := missing.Size(); err == nil || !isJavaIOException(err) {
		t.Fatalf("missing object error = %v", err)
	}
	if _, err := client.Put(7); err != nil {
		t.Fatal(err)
	}
	if err := client.CloseConnection(); err != nil {
		t.Fatal(err)
	}
	if !storage.Contains(7) {
		t.Fatal("closing a client connection changed remote storage")
	}
	if _, err := client.Put(8); err == nil {
		t.Fatal("closed client accepted an insertion")
	}
	if storage.Contains(8) {
		t.Fatal("closed client changed storage")
	}
}

func TestFPSetNullBlocksRejectBeforeStorageAccess(t *testing.T) {
	// All these implementations inherit FPSet's source null-vector contract.
	// No storage initialization is needed: failure precedes any storage access.
	for _, backend := range []struct {
		name string
		set  FPSet
	}{
		{"memory", &MemFPSet{}},
		{"memory1", &MemFPSet1{}},
		{"memory2", &MemFPSet2{}},
		{"disk", &DiskFPSet{}},
		{"multi", &MultiFPSet{}},
		{"noop", &NoopFPSet{}},
	} {
		for _, operation := range []string{"put", "contains"} {
			t.Run(backend.name+"/"+operation, func(t *testing.T) {
				defer func() {
					if failure := recover(); failure == nil {
						t.Fatal("null batch accepted as an empty batch")
					} else if _, ok := failure.(*NullPointerException); !ok {
						t.Fatalf("null batch failure = %T, want NullPointerException", failure)
					}
				}()
				if operation == "put" {
					backend.set.PutBlock(nil)
				} else {
					backend.set.ContainsBlock(nil)
				}
			})
		}
	}
}
