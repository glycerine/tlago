package tlc

import "testing"

// Source final reporting dereferences its fingerprint manager before publishing
// the final count. No original method directly tests a missing native owner.
func TestDistributedFingerprintStatisticsRequireOwner(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)
	for _, server := range []*TLCServer{nil, {}} {
		requireCoordinatorNullFailure(t, func() error { server.fpSetSize(); return nil })
		requireCoordinatorNullFailure(t, func() error { server.ReportSuccess(); return nil })
	}
	if recorder.Recorded(ECTLCSuccess) || len(ToolIOGetAllMessages()) != 0 {
		t.Fatal("missing fingerprint owner fabricated a success report")
	}
}

func TestDistributedFinalFingerprintOwnerRetained(t *testing.T) {
	manager := NewDistributedFPSetManagerFromFPSet(NewMemFPSet())
	manager.Put(73)
	trace := NewTLCTrace(t.TempDir(), "Spec")
	defer trace.Close()
	server := NewTLCServer("Spec", "Spec", t.TempDir(), manager, NewMemStateQueue(), trace)
	server.FPSetManager = nil
	if server.GetFPSetManager() != manager || server.fpSetSize() != 1 {
		t.Fatal("final fingerprint reporting followed the replacement mirror")
	}
}
