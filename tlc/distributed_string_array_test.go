package tlc

import "testing"

// ModelValue.data can carry string arrays. No original Java method directly
// tests their transfer; these checks exercise the native gob/TCP graph boundary.
func TestDistributedModelStringArrays(t *testing.T) {
	checkDistributedModelArrays(t, []any{[]string{"alpha", "λ\x00x", "", "\xff"}})
}

func TestDistributedModelStringArraysWorkerRPC(t *testing.T) {
	checkDistributedModelArraysWorkerRPC(t, []any{[]string{"alpha", "λ\x00x", "", "\xff"}})
}

func TestDistributedModelStringArraysInvalidReference(t *testing.T) {
	for _, reference := range []int{-1, 2} {
		payload := &DistributedStatePayload{
			Roots: []int{1}, States: []DistributedStateNode{{Level: 1, Values: []int{1}}},
			Values:       []DistributedValueNode{{Kind: "model", DataKind: "stringArray", DataArray: reference}},
			StringArrays: [][]string{{"alpha"}},
		}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid string array reference %d accepted", reference)
		}
	}
}
