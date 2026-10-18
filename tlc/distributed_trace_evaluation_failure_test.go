package tlc

import (
	"strings"
	"testing"
)

// No original method directly tests errors escaping trace reconstruction or
// alias evaluation. Source stops printing at the failing operation.
func TestDistributedTracePrintingPropagatesEvaluationFailures(t *testing.T) {
	initTLCCheckerTest(t)
	for _, operation := range []string{"alias", "current", "transition"} {
		for _, kind := range []string{"io", "fatal"} {
			t.Run(operation+"/"+kind, func(t *testing.T) {
				captureFailoverToolIO(t, ToolIOTool)
				recorder := &MemoryRecorder{}
				AddMessageRecorder(recorder)
				defer RemoveMessageRecorder(recorder)
				var failure error = NewIOException("trace evaluation failed")
				if kind == "fatal" {
					failure = NewAssertionError("trace evaluation failed")
				}
				current, successor := checkerTestState(1), checkerTestState(2)
				current.level, successor.level = 1, 2
				tool := &Tool{Actions: []*Action{{}}}
				wantPrinted := 0
				switch operation {
				case "alias":
					tool.EvalAliasInfoFunc = func(*Tool, *TLCStateInfo, *TLCStateMut, func() []*TLCStateInfo) (*TLCStateInfo, error) {
						return nil, failure
					}
				case "current":
					current.level = 2
					successor = nil
					tool.GetStateFunc = func(*Tool, uint64, ...any) (*TLCStateInfo, error) { return nil, failure }
				case "transition":
					wantPrinted = 1
					tool.GetNextStatesFunc = func(*Tool, *Action, *TLCStateMut) (*StateVec, error) { return nil, failure }
				}
				trace := NewTLCTrace()
				trace.Tool = tool
				err := invokeDistributedServerOperation(func() error { trace.printTraceWithPrefix(current, successor, []*TLCStateInfo{}); return nil })
				if err != failure {
					t.Fatalf("trace evaluation failure discarded: %T/%v", err, err)
				}
				if len(recorder.Records(ECTLCBehaviorUpToThisPoint)) != 1 || len(recorder.Records(ECTLCStatePrint2)) != wantPrinted || recorder.Recorded(ECGeneral) {
					t.Fatal("trace failure changed earlier output or printed fallback states")
				}
			})
		}
	}
}

func TestDistributedTraceAliasFailurePreservesHandlerCatch(t *testing.T) {
	initTLCCheckerTest(t)
	for _, kind := range []string{"io", "fatal"} {
		t.Run(kind, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			recorder := &MemoryRecorder{}
			AddMessageRecorder(recorder)
			defer RemoveMessageRecorder(recorder)
			var failure error = NewIOException("alias escaped tool")
			if kind == "fatal" {
				failure = NewAssertionError("alias escaped tool")
			}
			current, successor := checkerTestState(1), checkerTestState(2)
			current.level, successor.level = 1, 2
			trace := NewTLCTrace(t.TempDir(), "Spec")
			defer trace.Close()
			if _, err := trace.WriteStateRecord(nil, current.FingerPrint(), current); err != nil {
				t.Fatal(err)
			}
			trace.Tool = &Tool{EvalAliasInfoFunc: func(*Tool, *TLCStateInfo, *TLCStateMut, func() []*TLCStateInfo) (*TLCStateInfo, error) {
				return nil, failure
			}}
			queue := NewMemStateQueue()
			waiter := make(chan struct{})
			server := &TLCServer{Trace: trace, StateQueue: queue, completionWaiter: waiter}
			thread := &TLCServerThread{Server: server}
			original := NewWorkerException("original worker violation", current, successor, false)
			err := invokeDistributedServerOperation(func() error { thread.handleRunError(original, queue); return nil })
			if !server.IsDone() || server.ErrState != current || server.LastError != original || !server.KeepCallStack || len(recorder.Records(ECTLCBehaviorUpToThisPoint)) != 1 || recorder.Recorded(ECTLCStatePrint2) {
				t.Fatal("trace alias failure changed preceding model-error mutation/output")
			}
			if kind == "fatal" {
				if err != failure || queue.finish.Load() || recorder.Recorded(ECGeneral) || server.completionWaiter != waiter {
					t.Fatal("fatal alias failure entered ordinary printing catch or completed queue")
				}
				select {
				case <-waiter:
					t.Fatal("fatal trace failure notified completion")
				default:
				}
			} else {
				records := recorder.Records(ECGeneral)
				if err != nil || !queue.finish.Load() || len(records) != 1 || len(records[0].Params) != 1 || !strings.Contains(records[0].Params[0], failure.Error()) || strings.Contains(records[0].Params[0], original.Error()) {
					t.Fatalf("ordinary alias failure lost printing diagnostic/completion: %v, records %v", err, records)
				}
				select {
				case <-waiter:
				default:
					t.Fatal("caught trace failure skipped completion notification")
				}
			}
		})
	}
}
