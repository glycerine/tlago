package tlc

import "testing"

// RecordValue.PrintTLCState carries its record and underlying state. No direct
// upstream test exercises transfer; these checks use the native graph codec.
func printStatePayloadStates(t *testing.T) ([]*TLCStateMut, *UniqueString) {
	t.Helper()
	oldVariables, oldTool, oldCount := stateVariables, stateTool, UniqueStringVariableCount()
	oldPolicy := statePreserveMetadata
	t.Cleanup(func() {
		stateVariables, stateTool = oldVariables, oldTool
		statePreserveMetadata = oldPolicy
		SetUniqueStringVariableCount(oldCount)
	})
	field := &UniqueString{s: "payloadField", tok: 717, loc: 0}
	extra := &UniqueString{s: "extra", tok: 718, loc: -1}
	stateVariables = []StateVariable{{Name: field}}
	stateTool = nil
	statePreserveMetadata = true
	SetUniqueStringVariableCount(1)
	record := NewRecordValue([]*UniqueString{field, extra}, []Value{NewIntValue(42), NewIntValue(7)}, true)
	first, second := record.ToState(), record.ToState()
	first.UID, first.level = 31, 4
	first.SetCached(1, record)
	if first.GetCached(1) != nil || first.cached != nil {
		t.Fatal("source print wrapper inherits the base no-op cache")
	}
	mutable := &TLCStateMut{level: 5, values: []Value{record}}
	mutable.SetCached(1, record) // A separate extended mutable state owns this cache.
	return []*TLCStateMut{first, second, mutable, first}, extra
}

func requirePrintStatePayloadGraph(t *testing.T, original, copied []*TLCStateMut, extra *UniqueString) {
	t.Helper()
	if len(copied) != 4 || copied[0] == original[0] || copied[0] != copied[3] || copied[0].printRecord == nil || copied[0].printRecord != copied[1].printRecord || copied[0].printRecord == original[0].printRecord || copied[2].printRecord != nil {
		t.Fatal("print state lost record sharing, mutable-state distinction or receiver isolation")
	}
	record := copied[0].printRecord
	if copied[0].GetCached(1) != nil || copied[0].cached != nil || record != copied[2].GetCached(1) || record != copied[2].values[0] || record.Values[0] != copied[0].values[0] || copied[0].UID != 31 || copied[0].Level() != 4 {
		t.Fatal("print state lost record/value/cache sharing or stored metadata")
	}
	want := "/\\ payloadField = 42\n/\\ extra = 7\n"
	if got := copied[0].String(); got != want {
		t.Fatalf("received print state = %q, want %q", got, want)
	}
	if got := copied[0].Lookup(extra); got == nil || got.(*IntValue).Val != 7 || copied[0].ContainsKey(extra) || !copied[0].ContainsKey(record.Names[1]) {
		t.Fatal("received print state lost extra-field lookup or source membership identity")
	}
	if copied[0].printState == nil || copied[0].printState == original[0].printState || copied[0].printState == copied[1].printState || copied[0].printState.UID != TLCStateInitUID || copied[0].printState.Level() != TLCStateInitLevel {
		t.Fatal("print state lost separate underlying ownership or metadata")
	}
	if copied[0].FingerPrint() != original[0].FingerPrint() {
		t.Fatal("print state changed its underlying state fingerprint")
	}
	// Source ToState compares UniqueString.equals(UniqueString), which uses
	// tokens. Received names are separate objects with the same tokens.
	rebound := record.ToState()
	if rebound.values[0] != copied[0].values[0] || rebound.FingerPrint() != copied[0].FingerPrint() {
		t.Fatal("received record could not bind the local spec variable by token")
	}
	record.Values[1] = NewIntValue(9)
	if copied[1].Lookup(extra).(*IntValue).Val != 9 || original[0].Lookup(extra).(*IntValue).Val != 7 {
		t.Fatal("print-record mutation lost sharing or changed sender storage")
	}
}

func TestDistributedPrintStatePayloadGraph(t *testing.T) {
	original, extra := printStatePayloadStates(t)
	requirePrintStatePayloadGraph(t, original, distributedPayloadRoundTrip(t, original), extra)
}

func TestDistributedPrintStatePayloadValidation(t *testing.T) {
	for _, payload := range []*DistributedStatePayload{
		{States: []DistributedStateNode{{PrintRecord: -1}}, Roots: []int{1}},
		{States: []DistributedStateNode{{PrintRecord: 1}}, Roots: []int{1}},
		{States: []DistributedStateNode{{PrintRecord: 1}}, Values: []DistributedValueNode{{Kind: "bool"}}, Roots: []int{1}},
	} {
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatal("invalid print-record reference or type accepted")
		}
	}
}

func TestDistributedPrintStatePayloadFormatNameIdentity(t *testing.T) {
	original, _ := printStatePayloadStates(t)
	name := original[0].printRecord.Names[0]
	formatName := UniqueStringOf("_format")
	record := NewRecordValue([]*UniqueString{name, formatName}, []Value{NewIntValue(42), NewStringValue("%s:%s;")}, false)
	state := record.ToState()
	if got := state.String(); got != "payloadField:42;" {
		t.Fatalf("local print format = %q", got)
	}
	// Source Arrays.asList(names).indexOf(_FORMAT) uses Object.equals.
	// UniqueString overloads equals(UniqueString) but does not override that
	// method, so newly received names do not match the static format object.
	copied := distributedPayloadRoundTrip(t, []*TLCStateMut{state})[0]
	want := "/\\ payloadField = 42\n/\\ _format = \"%s:%s;\"\n"
	if copied.printRecord.Names[1] == formatName || copied.String() != want {
		t.Fatalf("received format-name identity/output = %q, want %q", copied.String(), want)
	}
}

func TestWorkerRPCPrintStatePayloadGraph(t *testing.T) {
	worker := &rpcTestWorker{next: func(states []*TLCStateMut) (*NextStateResult, error) {
		if len(states) == 2 {
			return nil, NewWorkerException("print-state failure", states[0], states[1], true)
		}
		if len(states) != 4 || states[0].printRecord == nil || states[0].printRecord != states[1].printRecord {
			t.Error("worker request lost print-state record sharing")
		}
		return NewNextStateResult([]*StateVec{NewStateVecFrom(states)}, []*LongVec{NewLongVec()}, 1, 4), nil
	}}
	_, client := startWorkerRPC(t, worker)
	original, extra := printStatePayloadStates(t)
	result, err := client.GetNextStates(original)
	if err != nil || result == nil || len(result.NextStates) != 1 {
		t.Fatalf("native print-state result = %v/%v", result, err)
	}
	requirePrintStatePayloadGraph(t, original, result.NextStates[0].ToSlice(), extra)
	result, err = client.GetNextStates(original[:2])
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || !failure.KeepCallStack || failure.State1 == nil || failure.State2 == nil || failure.State1.printRecord == nil || failure.State1.printRecord != failure.State2.printRecord || failure.State1.String() != original[0].String() {
		t.Fatalf("worker failure lost print-state record context: %v", err)
	}
}
