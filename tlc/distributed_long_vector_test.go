package tlc

import (
	"fmt"
	"math"
	"sync/atomic"
	"testing"
)

// LongVec's source transfer writes active elements, retaining object identity
// but neither spare capacity nor backing-array aliases. No upstream test covers
// attached vectors or their sharing with NextStateResult fingerprint vectors.
func attachedLongVectorState() *TLCStateMut {
	vector := NewLongVecWithCapacity(17)
	vector.Add(math.MinInt64)
	vector.Add(math.MaxInt64)
	other := &LongVec{data: vector.data}
	data := []any{vector, vector, other, NewLongVecWithCapacity(19), (*LongVec)(nil), vector.data, map[any]any{vector: other}}
	return &TLCStateMut{level: 1, values: []Value{&ModelValue{Data: data}}}
}

func checkAttachedLongVectors(state *TLCStateMut) error {
	data := state.values[0].(*ModelValue).Data.([]any)
	vector, other := data[0].(*LongVec), data[2].(*LongVec)
	if vector != data[1] || vector == other || vector.Size() != 2 || cap(vector.data) != 2 || cap(other.data) != 2 {
		return fmt.Errorf("attached vector identity, count or active storage changed")
	}
	if vector.ElementAt(0) != math.MinInt64 || vector.ElementAt(1) != math.MaxInt64 || other.ElementAt(0) != math.MinInt64 {
		return fmt.Errorf("attached vector fingerprint bits changed")
	}
	if &vector.data[0] == &other.data[0] || &vector.data[0] == &data[5].([]int64)[0] {
		return fmt.Errorf("custom vector transfer retained backing-array aliases")
	}
	if empty := data[3].(*LongVec); empty == nil || empty.Size() != 0 || cap(empty.data) != 0 {
		return fmt.Errorf("empty vector retained spare capacity or became null")
	}
	if data[4].(*LongVec) != nil || data[6].(map[any]any)[vector] != other {
		return fmt.Errorf("typed null or attached vector map-key identity changed")
	}
	return nil
}

func TestDistributedAttachedLongVectors(t *testing.T) {
	sender := attachedLongVectorState()
	got := distributedPayloadRoundTrip(t, []*TLCStateMut{sender})[0]
	if err := checkAttachedLongVectors(got); err != nil {
		t.Fatal(err)
	}
	vector := got.values[0].(*ModelValue).Data.([]any)[0].(*LongVec)
	vector.Add(41)
	if got.values[0].(*ModelValue).Data.([]any)[1].(*LongVec).Size() != 3 || sender.values[0].(*ModelValue).Data.([]any)[0].(*LongVec).Size() != 2 {
		t.Fatal("receiver mutation lost shared object identity or reached sender")
	}
}

func TestWorkerRPCAttachedLongVectors(t *testing.T) {
	var fail atomic.Bool
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(states []*TLCStateMut) (*NextStateResult, error) {
		if err := checkAttachedLongVectors(states[0]); err != nil {
			return nil, err
		}
		if fail.Load() {
			return nil, NewWorkerException("attached long vector", states[0], states[0], true)
		}
		vector := states[0].values[0].(*ModelValue).Data.([]any)[0].(*LongVec)
		return NewNextStateResult([]*StateVec{NewStateVecFrom(states)}, []*LongVec{vector, vector}, 1, 2), nil
	}})
	sender := attachedLongVectorState()
	result, err := client.GetNextStates([]*TLCStateMut{sender})
	if err != nil {
		t.Fatal(err)
	}
	state := result.NextStates[0].At(0)
	if err := checkAttachedLongVectors(state); err != nil {
		t.Fatal(err)
	}
	vector := state.values[0].(*ModelValue).Data.([]any)[0].(*LongVec)
	if result.NextFingerprints[0] != vector || result.NextFingerprints[1] != vector {
		t.Fatal("result fingerprint vectors lost sharing with attached model data")
	}
	vector.Add(41)
	if result.NextFingerprints[0].Size() != 3 || sender.values[0].(*ModelValue).Data.([]any)[0].(*LongVec).Size() != 2 {
		t.Fatal("result vector mutation lost sharing or reached sender")
	}
	fail.Store(true)
	result, err = client.GetNextStates([]*TLCStateMut{sender})
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || failure.State1 != failure.State2 || !failure.KeepCallStack {
		t.Fatalf("attached vector error context changed: %v/%v", result, err)
	}
	if err := checkAttachedLongVectors(failure.State1); err != nil {
		t.Fatal(err)
	}
}

func TestDistributedAttachedLongVectorInvalidReferences(t *testing.T) {
	for _, id := range []int{-1, 2} {
		payload := &DistributedStatePayload{LongVectors: [][]int64{{41}}, Values: []DistributedValueNode{{Kind: "model", DataKind: "longVector", DataArray: id}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid attached vector reference %d accepted", id)
		}
		payload.Values = nil
		payload.ObjectKeyMaps = [][]DistributedObjectKeyMapEntry{{{Key: DistributedObjectDataNode{Kind: "longVector", Array: id}}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid attached vector key %d accepted", id)
		}
		if _, err := DecodeDistributedResult(&DistributedResultPayload{States: &DistributedStatePayload{LongVectors: [][]int64{{41}}}, FingerprintVectors: []int{id}}); err == nil {
			t.Fatalf("invalid result vector reference %d accepted", id)
		}
	}
}
