package tlc

import "testing"

// Source TLCTrace stores links/fingerprints in its RAF, not retained state
// objects. No original method directly tests this ownership boundary.
func TestDiskTraceDoesNotRetainStateGraph(t *testing.T) {
	trace := NewTLCTrace(t.TempDir(), "Spec")
	t.Cleanup(func() { _ = trace.Close() })
	initial := &TLCStateMut{UID: TLCStateInitUID, level: 1}
	if err := trace.WriteInitState(initial, 41); err != nil {
		t.Fatal(err)
	}
	successor := &TLCStateMut{UID: initial.UID, level: 2}
	if _, err := trace.WriteStateRecord(successor, 43, successor); err != nil {
		t.Fatal(err)
	}
	third := &TLCStateMut{UID: TLCStateInitUID, level: 3}
	if err := trace.WriteNextState(successor, 47, third, nil); err != nil {
		t.Fatal(err)
	}
	if len(trace.records) != 0 || len(trace.Records()) != 0 {
		t.Fatal("disk trace retains written state objects")
	}
	if initial.UID != 0 || successor.UID != 12 || third.UID != 24 || trace.GetLevelForReporting() != 3 {
		t.Fatal("disk trace lost locations or predecessor chain")
	}
	if err := trace.BeginChkpt(); err != nil {
		t.Fatal(err)
	}
	enumerator, err := trace.Elements()
	if err != nil {
		t.Fatal(err)
	}
	defer enumerator.Close()
	for _, want := range []struct {
		location int64
		fp       uint64
	}{{0, 41}, {12, 43}, {24, 47}} {
		location, err := enumerator.NextPos()
		if err != nil || location != want.location {
			t.Fatalf("disk location = %d/%v, want %d", location, err, want.location)
		}
		fp, err := enumerator.NextFP()
		if err != nil || fp != want.fp {
			t.Fatalf("disk fingerprint = %d/%v, want %d", fp, err, want.fp)
		}
	}
}
