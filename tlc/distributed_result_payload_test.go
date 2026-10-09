package tlc

import (
	"bytes"
	"encoding/gob"
	"math"
	"reflect"
	"testing"
)

// There is no upstream result serialization test. These checks exercise the
// new Go wire representation; the original smart-proxy tests remain separate.
func resultPayloadRoundTrip(t *testing.T, result *NextStateResult) *NextStateResult {
	t.Helper()
	payload, err := EncodeDistributedResult(result)
	if err != nil {
		t.Fatal(err)
	}
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(payload); err != nil {
		t.Fatal(err)
	}
	var wire DistributedResultPayload
	if err := gob.NewDecoder(&buffer).Decode(&wire); err != nil {
		t.Fatal(err)
	}
	got, err := DecodeDistributedResult(&wire)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestDistributedResultPayloadPartitions(t *testing.T) {
	value := NewTupleValue([]Value{NewIntValue(42)})
	state := &TLCStateMut{WorkerID: 3, UID: -1, level: 40000, values: []Value{value}}
	first := NewStateVec(17)
	first.Add(state).Add(nil)
	second := NewStateVecFrom([]*TLCStateMut{state, &TLCStateMut{values: []Value{value}}})
	fingerprints := NewLongVecWithCapacity(19)
	for _, fp := range []int64{math.MinInt64, -1, math.MaxInt64} {
		fingerprints.Add(fp)
	}
	input := NewNextStateResult([]*StateVec{first, nil, second, first, NewStateVec(23)}, []*LongVec{fingerprints, nil, NewLongVec(), fingerprints}, math.MaxInt64, -1)
	got := resultPayloadRoundTrip(t, input)
	if got.ComputationTime != math.MaxInt64 || got.StatesComputed != -1 || got.GetStatesComputedDelta() != -6 {
		t.Fatal("result counters changed")
	}
	if len(got.NextStates) != 5 || len(got.NextFingerprints) != 4 {
		t.Fatal("partition widths changed")
	}
	if got.NextStates[0] != got.NextStates[3] || got.NextStates[0] == first || got.NextStates[1] != nil || got.NextStates[4] == nil {
		t.Fatal("state vector aliases/nulls changed")
	}
	if got.NextStates[0].At(0) != got.NextStates[2].At(0) || got.NextStates[0].At(1) != nil {
		t.Fatal("shared/null state changed")
	}
	if got.NextStates[0].At(0).values[0] != got.NextStates[2].At(1).values[0] || got.NextStates[0].At(0).values[0] == value {
		t.Fatal("value sharing or receiver ownership changed")
	}
	if got.NextFingerprints[0] != got.NextFingerprints[3] || got.NextFingerprints[0] == fingerprints || got.NextFingerprints[1] != nil || got.NextFingerprints[2] == nil {
		t.Fatal("fingerprint vector aliases/nulls changed")
	}
	if !reflect.DeepEqual(got.NextFingerprints[0].data, fingerprints.data) {
		t.Fatal("fingerprint bits changed")
	}
	if cap(got.NextStates[0].states) != 2 || cap(got.NextFingerprints[0].data) != 3 || cap(got.NextStates[4].states) != 0 {
		t.Fatal("spare source vector capacity crossed the wire")
	}
	got.NextFingerprints[0].data[0] = 0
	if fingerprints.data[0] != math.MinInt64 {
		t.Fatal("decoded storage aliases sender")
	}
}

func TestDistributedResultPayloadNulls(t *testing.T) {
	if got := resultPayloadRoundTrip(t, nil); got != nil {
		t.Fatal("null result became non-null")
	}
	for _, statesNil := range []bool{false, true} {
		for _, fingerprintsNil := range []bool{false, true} {
			input := NewNextStateResult([]*StateVec{}, []*LongVec{}, 17, -1)
			if statesNil {
				input.NextStates = nil
			}
			if fingerprintsNil {
				input.NextFingerprints = nil
			}
			got := resultPayloadRoundTrip(t, input)
			if (got.NextStates == nil) != statesNil || (got.NextFingerprints == nil) != fingerprintsNil || got.ComputationTime != 17 || got.StatesComputed != -1 {
				t.Fatal("null/empty arrays or dummy worker counters changed")
			}
		}
	}
}

func TestDistributedResultPayloadValidation(t *testing.T) {
	cases := map[string]*DistributedResultPayload{
		"missing":                         nil,
		"null contradiction":              {Nil: true, StateVectors: []int{0}},
		"null states contradiction":       {StatesNil: true, StateVectors: []int{0}},
		"null fingerprints contradiction": {FingerprintsNil: true, FingerprintVectors: []int{0}},
		"missing graph":                   {},
		"negative state":                  {States: &DistributedStatePayload{StateVectors: [][]int{{-1}}}},
		"oversize state":                  {States: &DistributedStatePayload{StateVectors: [][]int{{1}}}},
		"unused roots":                    {States: &DistributedStatePayload{Roots: []int{0}}},
		"bad vector":                      {States: &DistributedStatePayload{}, StateVectors: []int{1}},
		"negative vector":                 {States: &DistributedStatePayload{}, StateVectors: []int{-1}},
		"bad fingerprint":                 {States: &DistributedStatePayload{}, FingerprintVectors: []int{1}},
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeDistributedResult(payload); err == nil {
				t.Fatal("malformed result accepted")
			}
		})
	}
}
