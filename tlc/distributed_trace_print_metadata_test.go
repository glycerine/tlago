package tlc

import (
	"strings"
	"testing"
)

// Upstream has no direct tests of printTrace's branch-specific UID/worker writes.
func TestDistributedTracePrintingMetadata(t *testing.T) {
	initTLCCheckerTest(t)
	for _, branch := range []string{"initial-transition", "empty-prefix", "with-prefix", "noninitial-transition", "prefix-transition"} {
		t.Run(branch, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			current, successor := checkerTestState(2), checkerTestState(3)
			current.UID, current.WorkerID, current.level = 91, 9, 2
			successor.UID, successor.WorkerID, successor.level = 92, 10, 3
			regeneratedCurrent, regeneratedSuccessor := checkerTestState(2), checkerTestState(3)
			regeneratedCurrent.UID, regeneratedCurrent.WorkerID = 43, 4
			regeneratedSuccessor.UID, regeneratedSuccessor.WorkerID = 47, 6
			var prefix []*TLCStateInfo
			tool := &Tool{Actions: []*Action{{Name: "Next"}}, GetStateFunc: func(*Tool, uint64, ...any) (*TLCStateInfo, error) { return NewTLCStateInfo(regeneratedCurrent), nil }}
			tool.GetNextStatesFunc = func(*Tool, *Action, *TLCStateMut) (*StateVec, error) {
				return NewStateVecFrom([]*TLCStateMut{regeneratedSuccessor}), nil
			}
			switch branch {
			case "initial-transition":
				current.level = 1
			case "empty-prefix":
				successor = nil
			case "with-prefix", "prefix-transition":
				if branch == "with-prefix" {
					successor = nil
				}
				prefix = []*TLCStateInfo{NewTLCStateInfo(checkerTestState(1))}
			}
			wantUID, wantWorker := int64(43), int16(4)
			observed := 0
			tool.EvalAliasInfoFunc = func(_ *Tool, info *TLCStateInfo, _ *TLCStateMut, _ func() []*TLCStateInfo) (*TLCStateInfo, error) {
				if info.State == regeneratedCurrent {
					observed++
					if branch == "with-prefix" || branch == "prefix-transition" {
						wantUID, wantWorker = current.UID, current.WorkerID
					}
					if info.State.UID != wantUID || info.State.WorkerID != wantWorker {
						t.Fatalf("current metadata before alias = %d/%d, want %d/%d", info.State.UID, info.State.WorkerID, wantUID, wantWorker)
					}
				}
				if info.State == regeneratedSuccessor {
					observed++
					uid, worker := int64(47), int16(6)
					if branch == "noninitial-transition" || branch == "prefix-transition" {
						uid, worker = successor.UID, successor.WorkerID
					}
					if info.State.UID != uid || info.State.WorkerID != worker {
						t.Fatalf("successor metadata before alias = %d/%d, want %d/%d", info.State.UID, info.State.WorkerID, uid, worker)
					}
				}
				return info, nil
			}
			trace := NewTLCTrace()
			trace.Tool = tool
			trace.printTraceWithPrefix(current, successor, prefix)
			wantObserved := 1
			if branch == "noninitial-transition" || branch == "prefix-transition" {
				wantObserved = 2
			}
			if observed != wantObserved {
				t.Fatalf("reconstructed alias calls %d, want %d", observed, wantObserved)
			}
			if current.UID != 91 || current.WorkerID != 9 {
				t.Fatal("printing changed original current metadata")
			}
		})
	}
}

func TestDistributedInitialTraceMissingTransition(t *testing.T) {
	initTLCCheckerTest(t)
	captureFailoverToolIO(t, ToolIOTool)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)
	trace := NewTLCTrace()
	trace.Tool = &Tool{}
	current, successor := checkerTestState(1), checkerTestState(2)
	current.level, successor.level = 1, 2
	err := invokeDistributedServerOperation(func() error { trace.printTraceWithPrefix(current, successor, nil); return nil })
	if _, ok := err.(*NullPointerException); !ok {
		t.Fatalf("missing initial transition must fail at printing: %T/%v", err, err)
	}
	if len(recorder.Records(ECTLCBehaviorUpToThisPoint)) != 1 || len(recorder.Records(ECTLCStatePrint2)) != 1 || recorder.Recorded(ECTLCFailedToRecoverInit) || recorder.Recorded(ECTLCBug) || recorder.Recorded(ECTLCStatePrint1) {
		t.Fatal("missing initial transition changed source partial output or entered fatal recovery branch")
	}
}

func TestDistributedInitialTraceMissingTransitionHandler(t *testing.T) {
	initTLCCheckerTest(t)
	captureFailoverToolIO(t, ToolIOTool)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)
	current, successor := checkerTestState(1), checkerTestState(2)
	current.level, successor.level = 1, 2
	trace := NewTLCTrace(t.TempDir(), "Spec")
	defer trace.Close()
	if _, err := trace.WriteStateRecord(nil, current.FingerPrint(), current); err != nil {
		t.Fatal(err)
	}
	trace.Tool = &Tool{}
	queue := NewMemStateQueue()
	waiter := make(chan struct{})
	server := &TLCServer{Trace: trace, StateQueue: queue, completionWaiter: waiter}
	thread := &TLCServerThread{Server: server}
	original := NewWorkerException("original worker violation", current, successor, false)
	err := invokeDistributedServerOperation(func() error { thread.handleRunError(original, queue); return nil })
	if err != nil || !server.IsDone() || server.ErrState != current || server.LastError != original || !server.KeepCallStack || !queue.finish.Load() {
		t.Fatalf("missing-transition handler mutations/completion: %v", err)
	}
	records := recorder.Records(ECGeneral)
	if len(records) != 1 || len(records[0].Params) != 1 || !strings.Contains(records[0].Params[0], "PrintInvariantViolationStateTraceState") || strings.Contains(records[0].Params[0], original.Error()) {
		t.Fatalf("missing ordinary printing diagnostic: %v", records)
	}
	if len(recorder.Records(ECTLCBehaviorUpToThisPoint)) != 1 || len(recorder.Records(ECTLCStatePrint2)) != 1 || recorder.Recorded(ECTLCFailedToRecoverInit) || recorder.Recorded(ECTLCBug) {
		t.Fatal("handler changed partial trace output")
	}
	select {
	case <-waiter:
	default:
		t.Fatal("ordinary printing failure did not notify completion")
	}
}
