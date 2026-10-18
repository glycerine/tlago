package tlc

import "testing"

// TLCWorker's cache can miss the same fingerprint twice after a slot eviction.
// Its TreeSet then keeps the first Holder for that fingerprint.
func TestDistributedWorkerKeepsFirstHolderAfterCacheEviction(t *testing.T) {
	storage := NewMemFPSet()
	worker, _ := managerOwnershipWorker(t, 0, NewDistributedFPSetManager(NewLocalFingerprintEndpoint(storage)))
	worker.unsorted = false
	first := &TLCStateMut{values: []Value{NewIntValue(1)}}
	firstFP := first.FingerPrint()
	if firstFP == 0 {
		t.Fatal("first fingerprint would hit the cache's zero-filled slot")
	}
	var evicting *TLCStateMut
	for value := int32(2); value < 100000; value++ {
		candidate := &TLCStateMut{values: []Value{NewIntValue(value)}}
		fp := candidate.FingerPrint()
		if fp != firstFP && fp&1023 == firstFP&1023 {
			evicting = candidate
			break
		}
	}
	if evicting == nil {
		t.Fatal("could not find a second fingerprint for the cache slot")
	}
	second := &TLCStateMut{values: []Value{NewIntValue(1)}}
	parentOne := &TLCStateMut{UID: 31, level: 4}
	parentTwo := &TLCStateMut{UID: 41, level: 4}
	worker.App.Tool.GetNextStatesFunc = func(_ *Tool, _ *Action, parent *TLCStateMut) (*StateVec, error) {
		if parent == parentOne {
			return NewStateVecFrom([]*TLCStateMut{first, evicting}), nil
		}
		return NewStateVecFrom([]*TLCStateMut{second}), nil
	}

	result, err := worker.GetNextStates([]*TLCStateMut{parentOne, parentTwo})
	if err != nil || result == nil || result.StatesComputed != 3 {
		t.Fatalf("worker result = %v/%v", result, err)
	}
	if len(result.NextStates) != 1 || result.NextStates[0].Size() != 2 || result.NextFingerprints[0].Size() != 2 {
		t.Fatalf("duplicate fingerprint was not coalesced: %+v", result)
	}
	seenFirst := false
	for i := 0; i < 2; i++ {
		if uint64(result.NextFingerprints[0].ElementAt(i)) == firstFP {
			seenFirst = true
			if result.NextStates[0].At(i) != first || first.UID != parentOne.UID || second.UID == parentTwo.UID {
				t.Fatal("duplicate fingerprint replaced the first holder")
			}
		}
	}
	if !seenFirst || storage.Size() != 0 || worker.Cache.GetHitRate() != 1 {
		t.Fatal("worker changed cache misses or inserted fingerprints locally")
	}
}
