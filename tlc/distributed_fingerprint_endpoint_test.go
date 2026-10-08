package tlc

import "testing"

// Existing original failover tests throw local storage exceptions. Exercise
// returned endpoint errors as well, without changing the manager's algorithm.
func TestDistributedFingerprintEndpointReturnedFailure(t *testing.T) {
	for _, batch := range []bool{false, true} {
		t.Run(map[bool]string{false: "scalar", true: "batch"}[batch], func(t *testing.T) {
			failedStorage, spareStorage := NewMemFPSet(), NewMemFPSet()
			failed := &failedFingerprintEndpoint{
				LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(failedStorage),
				failure:                  NewIOException("endpoint unavailable"),
			}
			manager := NewDistributedFPSetManager(failed, NewLocalFingerprintEndpoint(spareStorage))
			if batch {
				fingerprints := []*LongVec{NewLongVec(), NewLongVec()}
				fingerprints[0].AddElement(2)
				result := manager.PutBlock(fingerprints)
				if result[0] == nil || !result[0].Get(0) {
					t.Fatal("new fingerprint lost its new-state answer after endpoint failure")
				}
			} else if manager.Put(2) {
				t.Fatal("new fingerprint reported as already present after endpoint failure")
			}
			if failed.calls != 1 || failedStorage.Contains(2) || !spareStorage.Contains(2) {
				t.Fatal("failed endpoint was retried or replacement did not receive the fingerprint")
			}
			if manager.NumOfAliveServers() != 1 || !manager.Contains(2) {
				t.Fatal("replacement availability or fingerprint lookup was not preserved")
			}
		})
	}
}

type failedFingerprintEndpoint struct {
	*LocalFingerprintEndpoint
	failure error
	calls   int
}

func TestDistributedFingerprintManagerWorkerSnapshot(t *testing.T) {
	manager := NewDynamicDistributedFPSetManager(3)
	for _, hostname := range []string{"first", "second", "third"} {
		if err := manager.RegisterFPSet(NewLocalFingerprintEndpoint(NewMemFPSet()), hostname); err != nil {
			t.Fatal(err)
		}
	}
	if manager.Reassign(0) != 1 {
		t.Fatal("first partition did not fail over")
	}
	manager.Trace = &TLCTrace{}
	endpoint := NewLocalServerEndpoint(&TLCServer{FPSetManager: manager})
	worker, err := endpoint.GetFPSetManager()
	if err != nil {
		t.Fatal(err)
	}
	if worker == manager || worker.Trace != nil || worker.GetMask() != manager.GetMask() || worker.NumOfAliveServers() != 2 {
		t.Fatal("snapshot shared manager state, included recovery trace, or lost partition identity")
	}
	if worker.Reassign(1) != 2 || worker.NumOfAliveServers() != 1 || manager.NumOfAliveServers() != 2 {
		t.Fatal("worker failover changed coordinator availability or lost shared-wrapper identity")
	}
	// The copy owns failover state, but fingerprint storage is still shared.
	if worker.Put(2) || !manager.Contains(2) {
		t.Fatal("snapshot duplicated fingerprint storage instead of retaining endpoints")
	}
	otherWorker, err := endpoint.GetFPSetManager()
	if err != nil || otherWorker == worker || otherWorker.NumOfAliveServers() != 2 {
		t.Fatal("later worker inherited another worker's failover")
	}
}

func (e *failedFingerprintEndpoint) Put(uint64) (bool, error) {
	e.calls++
	return false, e.failure
}

func (e *failedFingerprintEndpoint) PutBlock(*LongVec) (*BitVector, error) {
	e.calls++
	return nil, e.failure
}
