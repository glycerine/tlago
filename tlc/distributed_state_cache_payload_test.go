package tlc

import (
	"math"
	"strconv"
	"testing"
)

// The source state cache is a non-transient Map<Integer, Value>. No direct
// upstream transfer tests exist; native graph checks add no method credit.
func stateCachePayloadStates(t *testing.T) []*TLCStateMut {
	t.Helper()
	oldPolicy := statePreserveMetadata
	statePreserveMetadata = true // Only TLCStateMutExt owns a state cache.
	t.Cleanup(func() { statePreserveMetadata = oldPolicy })
	cycle := NewTupleValue(make([]Value, 1))
	cycle.Elems[0] = cycle
	shared := map[int]Value{math.MinInt32: cycle, math.MaxInt32: nil}
	empty := make(map[int]Value)
	parent := &TLCStateMut{level: 2, cached: shared}
	return []*TLCStateMut{
		{level: 3, values: []Value{cycle}, cached: shared, pred: parent},
		parent,
		{level: 1, cached: map[int]Value{math.MinInt32: cycle, math.MaxInt32: nil}},
		{level: 1, cached: empty}, {level: 1, cached: empty},
		{level: 1, cached: make(map[int]Value)}, {level: 1},
	}
}

func requireStateCachePayloadGraph(t *testing.T, original, copied []*TLCStateMut) {
	t.Helper()
	if len(copied) != 7 || copied[0] == original[0] || copied[0].Predecessor() != copied[1] {
		t.Fatal("cache transfer lost state graph ownership")
	}
	cycle, ok := copied[0].GetCached(math.MinInt32).(*TupleValue)
	if !ok || cycle == original[0].values[0] || cycle.Elems[0] != cycle || copied[0].values[0] != cycle || copied[1].GetCached(math.MinInt32) != cycle || copied[2].GetCached(math.MinInt32) != cycle {
		t.Fatal("cached value lost sharing, recursive identity or receiver isolation")
	}
	if value, present := copied[0].cached[math.MaxInt32]; !present || value != nil {
		t.Fatal("present null cache value became an absent key")
	}
	for _, index := range []int{3, 4, 5} {
		if copied[index].cached == nil || len(copied[index].cached) != 0 {
			t.Fatal("empty cache became nil or populated")
		}
	}
	if copied[6].cached != nil {
		t.Fatal("nil cache became empty")
	}
	copied[0].SetCached(17, NewIntValue(19))
	if copied[1].GetCached(17) != copied[0].GetCached(17) || copied[2].GetCached(17) != nil || original[0].GetCached(17) != nil {
		t.Fatal("shared cache mutation lost identity, merged equal maps or touched sender")
	}
	copied[3].SetCached(23, cycle)
	if copied[4].GetCached(23) != cycle || copied[5].GetCached(23) != nil || original[3].GetCached(23) != nil {
		t.Fatal("shared empty cache lost identity or merged separate empty maps")
	}
}

func TestDistributedStateCachePayloadGraph(t *testing.T) {
	original := stateCachePayloadStates(t)
	requireStateCachePayloadGraph(t, original, distributedPayloadRoundTrip(t, original))
}

func TestDistributedStateCachePayloadValidation(t *testing.T) {
	for _, reference := range []int{-1, 2} {
		payload := &DistributedStatePayload{States: []DistributedStateNode{{Cache: reference}}, StateCaches: [][]DistributedStateCacheEntry{{}}, Roots: []int{1}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid cache reference %d accepted", reference)
		}
	}
	for _, entries := range [][]DistributedStateCacheEntry{
		{{Key: 1, Value: -1}}, {{Key: 1, Value: 1}}, {{Key: 1}, {Key: 1}},
	} {
		payload := &DistributedStatePayload{States: []DistributedStateNode{{Cache: 1}}, StateCaches: [][]DistributedStateCacheEntry{entries}, Roots: []int{1}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid cache entries accepted: %#v", entries)
		}
	}
	if strconv.IntSize == 64 {
		for _, key := range []int64{math.MinInt32 - 1, math.MaxInt32 + 1} {
			state := &TLCStateMut{cached: map[int]Value{int(key): nil}}
			if _, err := EncodeDistributedStates([]*TLCStateMut{state}); err == nil {
				t.Fatalf("out-of-range cache key %d accepted", key)
			}
		}
	}
}

func TestWorkerRPCStateCachePayloadGraph(t *testing.T) {
	worker := &rpcTestWorker{next: func(states []*TLCStateMut) (*NextStateResult, error) {
		if len(states) == 2 {
			return nil, NewWorkerException("cached state failure", states[0], states[1], true)
		}
		if len(states) != 7 || states[0].GetCached(math.MinInt32) != states[0].values[0] || states[0].Predecessor() != states[1] {
			t.Error("worker request lost cached state graph")
		}
		return NewNextStateResult([]*StateVec{NewStateVecFrom(states)}, []*LongVec{NewLongVec()}, 1, 7), nil
	}}
	_, client := startWorkerRPC(t, worker)
	original := stateCachePayloadStates(t)
	result, err := client.GetNextStates(original)
	if err != nil || result == nil || len(result.NextStates) != 1 {
		t.Fatalf("native cache graph result = %v/%v", result, err)
	}
	requireStateCachePayloadGraph(t, original, result.NextStates[0].ToSlice())
	result, err = client.GetNextStates(original[:2])
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || !failure.KeepCallStack || failure.State1 == nil || failure.State2 == nil || failure.State1.Predecessor() != failure.State2 {
		t.Fatalf("cached worker failure context = %v/%v", result, err)
	}
	if failure.State1.GetCached(math.MinInt32) != failure.State2.GetCached(math.MinInt32) || failure.State1.GetCached(math.MinInt32) != failure.State1.values[0] {
		t.Fatal("worker failure cache lost shared values")
	}
	failure.State1.SetCached(31, NewIntValue(37))
	if failure.State2.GetCached(31) != failure.State1.GetCached(31) || original[0].GetCached(31) != nil {
		t.Fatal("worker failure cache lost map sharing or receiver isolation")
	}
}
