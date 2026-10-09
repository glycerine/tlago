package tlc

import "testing"

// No direct original test covers toRecordValue's PrintTLCState branch.
func TestStateInfoPreservesPrintRecordIdentity(t *testing.T) {
	states, _ := printStatePayloadStates(t)
	for _, wrapper := range []*TLCStateMut{states[0], distributedPayloadRoundTrip(t, states[:1])[0]} {
		record := wrapper.printRecord
		info := NewTLCStateInfo(wrapper)
		if info.ToRecordValue() != record {
			t.Fatal("state info rebuilt the print record instead of returning it")
		}
		counterexample := NewCounterExampleFromTrace([]*TLCStateInfo{info})
		stateSet, err := counterexample.Select(NewStringValue("state"))
		if err != nil {
			t.Fatal(err)
		}
		node := stateSet.(*SetEnumValue).Elems.At(0).(*TupleValue)
		if node.Elems[1] != record {
			t.Fatal("counterexample copied the print record")
		}
		field := record.Names[0]
		wrapper.Bind(field, NewIntValue(43))
		ordinary := NewTLCStateInfo(wrapper.printState).ToRecordValue().(*RecordValue)
		if ordinary == record || len(ordinary.Names) != len(stateVariables) || ordinary.Values[0].(*IntValue).Val != 43 || ordinary.IsNorm {
			t.Fatal("ordinary state info did not construct a fresh variable record")
		}
		record.Values[0] = NewIntValue(44)
		if node.Elems[1].(*RecordValue).Values[0].(*IntValue).Val != 44 || ordinary.Values[0].(*IntValue).Val != 43 {
			t.Fatal("record identity did not preserve mutation visibility and snapshot distinction")
		}
	}
}

func TestAliasTLCStateInfoPreservesOriginalStateActionAndLevel(t *testing.T) {
	initTLCCheckerTest(t)
	SetTLCStateTool(NewTool().SetMode(ModeSimulation))
	t.Cleanup(func() { SetTLCStateTool(nil) })
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
