package tlc

import "testing"

// No upstream method directly tests these reconstruction failure boundaries.
func TestConcurrentTraceReconstruction(t *testing.T) {
	initTLCCheckerTest(t)
	for _, scenario := range []string{"empty-records", "anchor-only", "success", "provided-initial", "initial-error", "successor-error", "missing-initial"} {
		t.Run(scenario, func(t *testing.T) {
			prior := NewJavaRandom(123)
			prior.enumerableKind = randomEnumerableDefault
			old := SetRandomEnumerableGenerator(prior)
			defer SetRandomEnumerableGenerator(old)
			initial, successor := NewTLCStateInfo(checkerTestState(1)), NewTLCStateInfo(checkerTestState(2))
			initial.State.UID, initial.State.WorkerID = 77, 7
			successor.State.UID, successor.State.WorkerID = 88, 8
			records := []ConcurrentTraceRecord{{Ptr: 22, Worker: 65539}, {Ptr: 11, Worker: 4, FP: 23}, {Ptr: 1, FP: 17}}
			var supplied *TLCStateInfo
			switch scenario {
			case "empty-records":
				records = nil
			case "anchor-only":
				records = records[2:]
			case "provided-initial":
				supplied = initial
			}
			failure := NewIOException("concurrent reconstruction failed")
			calls := 0
			tool := &Tool{GetStateFunc: func(_ *Tool, fp uint64, prev ...any) (*TLCStateInfo, error) {
				calls++
				if RandomEnumerableGenerator() == prior {
					t.Fatal("reconstruction did not reset randomness")
				}
				if len(prev) == 0 {
					if fp != 17 {
						t.Fatalf("initial lookup %d", fp)
					}
					if scenario == "initial-error" {
						return nil, failure
					}
					if scenario == "missing-initial" {
						return nil, nil
					}
					return initial, nil
				}
				if len(prev) != 1 || prev[0] != initial.State || fp != 23 {
					t.Fatalf("successor lookup %d/%v", fp, prev)
				}
				uid, worker := int64(11), int16(4)
				if scenario == "provided-initial" {
					uid, worker = 77, 7
				}
				if initial.State.UID != uid || initial.State.WorkerID != worker {
					t.Fatal("initial metadata not set before successor lookup")
				}
				if scenario == "successor-error" {
					return nil, failure
				}
				return successor, nil
			}}
			trace := &ConcurrentTLCTrace{Tool: tool}
			var got []*TLCStateInfo
			err := invokeDistributedServerOperation(func() error { var e error; got, e = trace.recoverTraceFromRecords(supplied, records); return e })
			switch scenario {
			case "empty-records":
				if _, ok := err.(*NegativeArraySizeException); !ok || calls != 0 || RandomEnumerableGenerator() == prior {
					t.Fatalf("empty-record allocation failure: %T/%v", err, err)
				}
			case "initial-error", "successor-error":
				if err != failure || RandomEnumerableGenerator() == prior {
					t.Fatalf("error/random lifetime: %T/%v", err, err)
				}
				if scenario == "successor-error" && (initial.State.UID != 11 || initial.State.WorkerID != 4) {
					t.Fatal("failure lost preceding initial metadata writes")
				}
			case "missing-initial":
				if _, ok := err.(*NullPointerException); !ok || calls != 1 || RandomEnumerableGenerator() == prior {
					t.Fatalf("missing initial dereference: %T/%v, calls %d", err, err, calls)
				}
			case "anchor-only":
				if err != nil || got == nil || len(got) != 0 || calls != 0 || RandomEnumerableGenerator() != prior {
					t.Fatalf("anchor-only result/calls/random: %v/%v/%d", got, err, calls)
				}
			default:
				if err != nil || len(got) != 2 || got[0] != initial || got[1] != successor || RandomEnumerableGenerator() != prior {
					t.Fatalf("normal result/random: %v/%v", got, err)
				}
				if successor.State.UID != 22 || successor.State.WorkerID != 3 {
					t.Fatalf("successor metadata: %d/%d", successor.State.UID, successor.State.WorkerID)
				}
			}
		})
	}
}

func TestConcurrentTracePublicRecoveryFailure(t *testing.T) {
	initTLCCheckerTest(t)
	for _, operation := range []string{"from-state", "between-states", "print"} {
		for _, kind := range []string{"io", "fatal"} {
			t.Run(operation+"/"+kind, func(t *testing.T) {
				captureFailoverToolIO(t, ToolIOTool)
				recorder := &MemoryRecorder{}
				AddMessageRecorder(recorder)
				defer RemoveMessageRecorder(recorder)
				worker := NewWorker(0)
				worker.SetTraceContext(t.TempDir(), "Spec")
				defer worker.CloseTrace()
				initial, middle, end := checkerTestState(1), checkerTestState(2), checkerTestState(3)
				if err := worker.ensureTraceRAF(); err != nil {
					t.Fatal(err)
				}
				if err := worker.WriteInitState(initial, initial.FingerPrint()); err != nil {
					t.Fatal(err)
				}
				if err := worker.WriteNextState(initial, middle.FingerPrint(), middle, nil); err != nil {
					t.Fatal(err)
				}
				if err := worker.WriteNextState(middle, end.FingerPrint(), end, nil); err != nil {
					t.Fatal(err)
				}
				var failure error = NewIOException("public trace recovery failed")
				if kind == "fatal" {
					failure = NewAssertionError("public trace recovery failed")
				}
				trace := &ConcurrentTLCTrace{TLCTrace: NewTLCTrace(), Workers: []*Worker{worker}}
				trace.SetTool(&Tool{GetStateFunc: func(*Tool, uint64, ...any) (*TLCStateInfo, error) {
					if trace.mu.TryLock() {
						trace.mu.Unlock()
						t.Error("reconstruction escaped the trace monitor")
					}
					return nil, failure
				}})
				err := invokeDistributedServerOperation(func() error {
					switch operation {
					case "from-state":
						trace.GetTraceFromState(end)
					case "between-states":
						trace.GetTraceBetweenStates(initial, end)
					case "print":
						trace.PrintTrace(end, nil)
					}
					return nil
				})
				if err != failure {
					t.Fatalf("public recovery failure identity: %T/%v", err, err)
				}
				if !trace.mu.TryLock() {
					t.Fatal("failed reconstruction retained the trace monitor")
				}
				trace.mu.Unlock()
				if recorder.Recorded(ECTLCBehaviorUpToThisPoint) || recorder.Recorded(ECTLCStatePrint2) {
					t.Fatal("failed prefix recovery entered trace printing")
				}
			})
		}
	}
}
