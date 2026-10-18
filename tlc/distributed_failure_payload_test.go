package tlc

import (
	"bytes"
	"encoding/gob"
	"testing"
)

// Native endpoint fixtures set the same operation traits as Go adapters, without
// fabricating Java remote exceptions. Tests retain failure identity and text.
func distributedTestRemoteFailure(message string) *DistributedOperationError {
	return newDistributedOperationError(DistributedOperationError{Message: javaString(message), Class: "tlc.DistributedOperationFailure", Remote: true, IO: true})
}

func TestDistributedFailurePayloadCyclesAndOwnership(t *testing.T) {
	failure := &DistributedOperationError{Message: javaString("cyclic cause"), Class: "native failure"}
	failure.Cause = failure
	failure.Suppressed = []error{failure}
	payload, err := EncodeDistributedFailure(failure)
	if err != nil {
		t.Fatal(err)
	}
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(payload); err != nil {
		t.Fatal(err)
	}
	var wire DistributedFailurePayload
	if err := gob.NewDecoder(&buffer).Decode(&wire); err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeDistributedFailure(&wire)
	if err != nil {
		t.Fatal(err)
	}
	got := decoded.(*DistributedOperationError)
	if got == failure || got.Cause != got || len(got.Suppressed) != 1 || got.Suppressed[0] != got {
		t.Fatal("cause/suppression cycle or ownership changed")
	}
	if got.Message == failure.Message || got.Error() != failure.Error() || got.Class != failure.Class {
		t.Fatal("diagnostic ownership changed")
	}
}

func TestDistributedFailurePayloadValidation(t *testing.T) {
	cases := map[string]*DistributedFailurePayload{
		"missing":            nil,
		"root":               {Root: 1},
		"null contradiction": {Nodes: []DistributedFailureNode{{}}},
		"missing states":     {Root: 1, Nodes: []DistributedFailureNode{{}}},
		"cause":              {Root: 1, Nodes: []DistributedFailureNode{{Cause: 2}}, States: &DistributedStatePayload{}},
		"suppressed":         {Root: 1, Nodes: []DistributedFailureNode{{Suppressed: []int{0}}}, States: &DistributedStatePayload{}},
		"worker states":      {Root: 1, Nodes: []DistributedFailureNode{{Worker: true}}, States: &DistributedStatePayload{}},
		"non-worker states":  {Root: 1, Nodes: []DistributedFailureNode{{State1: 1}}, States: &DistributedStatePayload{}},
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeDistributedFailure(payload); err == nil {
				t.Fatal("malformed failure accepted")
			}
		})
	}
	payload, err := EncodeDistributedFailure(nil)
	if err != nil {
		t.Fatal(err)
	}
	if failure, err := DecodeDistributedFailure(payload); err != nil || failure != nil {
		t.Fatalf("null failure: %v/%v", failure, err)
	}
}
