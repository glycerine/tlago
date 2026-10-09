package tlc

import (
	"fmt"
	"math"
	"sync/atomic"
	"testing"
)

// BitVector has default source field transfer, unlike LongVec's custom active
// prefix transfer. Preserve its complete word array and shared storage.
func attachedBitVectorState() *TLCStateMut {
	words := []uint64{1, uint64(1) << 63, math.MaxUint64, 0, 0}
	vector := &BitVector{word: words}
	other := &BitVector{word: words}
	distinct := &BitVector{word: append([]uint64(nil), words...)}
	data := []any{vector, vector, other, distinct, words, &BitVector{}, NewBitVector(0), (*BitVector)(nil),
		map[any]any{vector: other}, []uint64(nil), make([]uint64, 0)}
	return &TLCStateMut{level: 1, values: []Value{&ModelValue{Data: data}}}
}

func checkAttachedBitVectors(state *TLCStateMut) error {
	data := state.values[0].(*ModelValue).Data.([]any)
	vector, other, distinct := data[0].(*BitVector), data[2].(*BitVector), data[3].(*BitVector)
	if vector != data[1] || vector == other || vector == distinct || other == distinct || len(vector.word) != 5 {
		return fmt.Errorf("bit-vector identities or complete storage length changed")
	}
	if vector.word[0] != 1 || vector.word[1] != uint64(1)<<63 || vector.word[2] != math.MaxUint64 || vector.word[3] != 0 || vector.word[4] != 0 {
		return fmt.Errorf("bit-vector word bits changed")
	}
	if &vector.word[0] != &other.word[0] || &vector.word[0] != &data[4].([]uint64)[0] || &vector.word[0] == &distinct.word[0] {
		return fmt.Errorf("bit-vector backing storage identity changed")
	}
	if data[5].(*BitVector).word != nil || data[6].(*BitVector).word == nil || len(data[6].(*BitVector).word) != 0 || data[7].(*BitVector) != nil {
		return fmt.Errorf("null/empty vector or word storage changed")
	}
	if data[8].(map[any]any)[vector] != other || data[9].([]uint64) != nil || data[10].([]uint64) == nil {
		return fmt.Errorf("bit-vector map key or null/empty word attachment changed")
	}
	return nil
}

func TestDistributedAttachedBitVectors(t *testing.T) {
	sender := attachedBitVectorState()
	got := distributedPayloadRoundTrip(t, []*TLCStateMut{sender})[0]
	if err := checkAttachedBitVectors(got); err != nil {
		t.Fatal(err)
	}
	data := got.values[0].(*ModelValue).Data.([]any)
	data[0].(*BitVector).Set(257)
	if !data[2].(*BitVector).Get(257) || data[4].([]uint64)[4] != 2 || data[3].(*BitVector).Get(257) || sender.values[0].(*ModelValue).Data.([]any)[0].(*BitVector).Get(257) {
		t.Fatal("receiver mutation lost sharing, merged distinct storage or reached sender")
	}
}

func TestWorkerRPCAttachedBitVectors(t *testing.T) {
	var fail atomic.Bool
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(states []*TLCStateMut) (*NextStateResult, error) {
		if err := checkAttachedBitVectors(states[0]); err != nil {
			return nil, err
		}
		if fail.Load() {
			return nil, NewWorkerException("attached bit vector", states[0], states[0], true)
		}
		return NewNextStateResult([]*StateVec{NewStateVecFrom(states)}, []*LongVec{NewLongVec()}, 1, 2), nil
	}})
	sender := attachedBitVectorState()
	result, err := client.GetNextStates([]*TLCStateMut{sender})
	if err != nil {
		t.Fatal(err)
	}
	if err := checkAttachedBitVectors(result.NextStates[0].At(0)); err != nil {
		t.Fatal(err)
	}
	data := result.NextStates[0].At(0).values[0].(*ModelValue).Data.([]any)
	data[0].(*BitVector).Set(257)
	if !data[2].(*BitVector).Get(257) || data[4].([]uint64)[4] != 2 || data[3].(*BitVector).Get(257) || sender.values[0].(*ModelValue).Data.([]any)[0].(*BitVector).Get(257) {
		t.Fatal("TCP receiver mutation lost sharing or isolation")
	}
	fail.Store(true)
	result, err = client.GetNextStates([]*TLCStateMut{sender})
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || failure.State1 != failure.State2 || !failure.KeepCallStack {
		t.Fatalf("bit-vector failure state context changed: %v/%v", result, err)
	}
	if err := checkAttachedBitVectors(failure.State1); err != nil {
		t.Fatal(err)
	}
}

func TestDistributedAttachedBitVectorInvalidReferences(t *testing.T) {
	for _, id := range []int{-1, 2} {
		payload := &DistributedStatePayload{BitVectors: []int{0}, Values: []DistributedValueNode{{Kind: "model", DataKind: "bitVector", DataArray: id}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid bit-vector object reference %d accepted", id)
		}
		payload.Values = nil
		payload.ObjectKeyMaps = [][]DistributedObjectKeyMapEntry{{{Key: DistributedObjectDataNode{Kind: "bitVector", Array: id}}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid bit-vector map-key reference %d accepted", id)
		}
		payload = &DistributedStatePayload{BitVectors: []int{id}, PrimitiveArrays: []DistributedPrimitiveArrayNode{{Kind: "uint64Array", Bits: []uint64{1}}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid bit-vector word reference %d accepted", id)
		}
	}
	payload := &DistributedStatePayload{BitVectors: []int{1}, PrimitiveArrays: []DistributedPrimitiveArrayNode{{Kind: "int64Array", Bits: []uint64{1}}}}
	if _, err := DecodeDistributedStates(payload); err == nil {
		t.Fatal("bit-vector word storage accepted the wrong native array type")
	}
}
