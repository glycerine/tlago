package tlc

import "testing"

func TestAliasTLCStateInfoPreservesOriginalStateActionAndLevel(t *testing.T) {
	initTLCCheckerTest(t)
	originalAction := &Action{Name: "OriginalNext"}
	initState := checkerTestState(0)
	original := checkerTestState(1).SetPredecessor(initState).SetAction(originalAction)
	alias := checkerTestState(99)

	current := NewTLCStateInfo(original)
	info := AliasTLCStateInfo(alias, current)

	if info.State != alias {
		t.Fatalf("alias info State = %p, want alias %p", info.State, alias)
	}
	if info.OriginalState() != original {
		t.Fatalf("alias original state = %p, want original %p", info.OriginalState(), original)
	}
	if info.GetStateNumber() != original.Level() {
		t.Fatalf("alias state number = %d, want original level %d", info.GetStateNumber(), original.Level())
	}
	if info.Action() != originalAction {
		t.Fatalf("alias action = %v, want original action %v", info.Action(), originalAction)
	}
}

func TestTLCStateInfoCachesFingerprintAndComparesByState(t *testing.T) {
	initTLCCheckerTest(t)
	state := checkerTestState(1)
	info := NewTLCStateInfo(state)
	fp := info.FingerPrint()
	state.Bind(UniqueStringOf("x"), NewIntValue(2))
	if got := info.FingerPrint(); got != fp {
		t.Fatalf("cached fingerprint = %d, want original %d", got, fp)
	}
	if !info.Equal(state) {
		t.Fatalf("state info should compare equal to its current state object")
	}
	if info.Equal(checkerTestState(1)) {
		t.Fatalf("state info compared equal to a different state after mutation")
	}
}
