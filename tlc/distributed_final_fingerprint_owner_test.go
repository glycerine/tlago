package tlc

import (
	"os"
	"os/exec"
	"testing"
)

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

func TestDistributedFinalFingerprintOwnerFailureProcess(t *testing.T) {
	if os.Getenv("TLAGO_FINAL_FP_OWNER_FAILURE") == "" {
		command := exec.Command(os.Args[0], "-test.run=^TestDistributedFinalFingerprintOwnerFailureProcess$")
		command.Env = append(os.Environ(), "TLAGO_FINAL_FP_OWNER_FAILURE=true", tlcServerPropertyPrefix+".report=1")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("final fingerprint ownership process failed: %v\n%s", err, output)
		}
		return
	}
	captureFailoverToolIO(t, ToolIOTool)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)
	directory := t.TempDir()
	trace := NewTLCTrace(directory, "Spec")
	defer trace.Close()
	server := NewTLCServer("Spec", "Spec", directory, NewDistributedFPSetManager(), NewMemStateQueue(directory), trace)
	server.SetTool(NewTool())
	server.StatesPerMinute, server.DistinctStatesPerMinute = 17, 19
	lateCalls := 0
	server.ConfigurePublication(TLCServerPublication{
		LocalHostName: func() (string, error) { return "local", nil },
		CreateRegistry: func(int) (*TLCServerRegistry, error) {
			return &TLCServerRegistry{
				Rebind: func(name string, s *TLCServer) error {
					if name == TLCServerWorkerName {
						s.SetDone()
						s.FPSetManager = nil
					}
					return nil
				},
				Unbind: func(string) error { lateCalls++; return nil },
			}, nil
		},
		Flush:    func() { lateCalls++ },
		Unexport: func(*TLCServer, bool) (bool, error) { lateCalls++; return true, nil },
	})
	previous := TLCServerFinalNumberOfDistinctStates()
	err := invokeDistributedServerOperation(func() error { _, err := server.ModelCheck(); return err })
	if _, ok := err.(*NullPointerException); !ok {
		t.Fatalf("missing final manager: %T/%v", err, err)
	}
	if !server.executor.IsShutdown() || server.FinalNumberOfDistinctStates != -1 || TLCServerFinalNumberOfDistinctStates() != previous {
		t.Fatal("missing manager changed earlier executor shutdown or published a fabricated final count")
	}
	if server.StatesPerMinute != 17 || server.DistinctStatesPerMinute != 19 || lateCalls != 0 || recorder.Recorded(ECTLCSuccess) || recorder.Recorded(ECTLCStats) || recorder.Recorded(ECTLCFinished) {
		t.Fatal("missing final manager continued through later reporting or cleanup")
	}
	if _, err := os.Stat(directory); err != nil {
		t.Fatal("failed final reporting deleted metadata", err)
	}
}
