package tlc

import (
	"math"
	"testing"
)

// No source test directly covers these result getters.
func TestDistributedResultGetterNullContract(t *testing.T) {
	for _, call := range []func(){
		func() { (*NextStateResult)(nil).GetStatesComputedDelta() },
		func() { (*NextStateResult)(nil).GetComputationTime() },
		func() { (*NextStateResult)(nil).GetNextStates() },
		func() { (*NextStateResult)(nil).GetNextFingerprints() },
		func() { NewNextStateResult(nil, nil, 0, 5).GetStatesComputedDelta() },
	} {
		func() {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("null getter failure = %T/%v", failure, failure)
				}
			}()
			call()
		}()
	}
}

func TestDistributedResultGetterCountersAndReferences(t *testing.T) {
	states := []*StateVec{NewStateVec(0), NewStateVec(0)}
	fps := []*LongVec{NewLongVec()}
	result := NewNextStateResult(states, fps, math.MinInt64, math.MinInt64)
	if result.GetStatesComputedDelta() != math.MaxInt64-1 || result.GetComputationTime() != math.MinInt64 {
		t.Fatal("signed result counters changed")
	}
	if &result.GetNextStates()[0] != &states[0] || &result.GetNextFingerprints()[0] != &fps[0] {
		t.Fatal("getter copied the result arrays")
	}
	result = NewNextStateResult([]*StateVec{}, nil, 0, 5)
	if result.GetStatesComputedDelta() != 5 || result.GetNextStates() == nil || result.GetNextFingerprints() != nil {
		t.Fatal("empty and null result arrays were conflated")
	}
}
