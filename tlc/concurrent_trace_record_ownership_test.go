package tlc

import "testing"

// No original method directly exercises unavailable trace-worker owners.
func TestConcurrentTraceRequiresRecordWorker(t *testing.T) {
	initTLCCheckerTest(t)
	for _, scenario := range []string{"empty", "negative", "outside", "nil-slot"} {
		for _, operation := range []string{"from-state", "between-states", "print"} {
			t.Run(scenario+"/"+operation, func(t *testing.T) {
				captureFailoverToolIO(t, ToolIOTool)
				recorder := &MemoryRecorder{}
				AddMessageRecorder(recorder)
				defer RemoveMessageRecorder(recorder)
				initial, end := checkerTestState(1), checkerTestState(2)
				end.SetPredecessor(initial)
				end.WorkerID = 0
				trace := &ConcurrentTLCTrace{TLCTrace: NewTLCTrace(), Tool: &Tool{}, Workers: []*Worker{nil}}
				switch scenario {
				case "empty":
					trace.Workers = nil
				case "negative":
					end.WorkerID = -1
				case "outside":
					end.WorkerID = 1
				}
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
				if scenario == "nil-slot" {
					if _, ok := err.(*NullPointerException); !ok {
						t.Fatalf("required slot failure: %T/%v", err, err)
					}
				} else if _, ok := err.(*ArrayIndexOutOfBoundsException); !ok {
					t.Fatalf("required index failure: %T/%v", err, err)
				}
				if recorder.Recorded(ECTLCBehaviorUpToThisPoint) || recorder.Recorded(ECTLCStatePrint2) {
					t.Fatal("unavailable record printed fallback trace")
				}
			})
		}
	}
}

func TestConcurrentTracePredecessorRequiresWorker(t *testing.T) {
	for _, index := range []int{-1, 0, 1} {
		trace := &ConcurrentTLCTrace{}
		record := ConcurrentTraceRecord{Ptr: 37, Worker: index, Workers: []*Worker{nil}}
		err := invokeDistributedServerOperation(func() error { _, err := trace.predecessorRecord(record); return err })
		if index == 0 {
			if _, ok := err.(*NullPointerException); !ok {
				t.Fatalf("predecessor slot failure: %T/%v", err, err)
			}
		} else if _, ok := err.(*ArrayIndexOutOfBoundsException); !ok {
			t.Fatalf("predecessor index failure: %T/%v", err, err)
		}
	}
}

func TestConcurrentTraceInitialPredecessorStillReadsRecord(t *testing.T) {
	worker := NewWorker(0)
	worker.SetTraceContext(t.TempDir(), "Spec")
	defer worker.CloseTrace()
	if err := worker.ensureTraceRAF(); err != nil {
		t.Fatal(err)
	}
	if err := worker.traceRAF.Close(); err != nil {
		t.Fatal(err)
	}
	record := ConcurrentTraceRecord{Ptr: 1, Workers: []*Worker{worker}}
	_, err := (&ConcurrentTLCTrace{}).predecessorRecord(record)
	if !isJavaIOException(err) {
		t.Fatalf("initial predecessor skipped record access: %v", err)
	}
}

func TestConcurrentTraceInitialAndEqualDoNotRequireWorkers(t *testing.T) {
	initTLCCheckerTest(t)
	trace := &ConcurrentTLCTrace{}
	initial := checkerTestState(1)
	end := checkerTestState(2)
	end.level = 2
	if got := trace.GetTraceFromState(initial); len(got) != 1 || got[0].State != initial {
		t.Fatal("initial short path required workers")
	}
	if got := trace.GetTraceBetweenStates(nil, initial); len(got) != 1 || got[0].State != initial {
		t.Fatal("initial target required from/worker")
	}
	if got := trace.GetTraceBetweenStates(end, end); len(got) != 1 || got[0].State != end {
		t.Fatal("equal-state short path required workers")
	}
}

func TestConcurrentTraceCollectsRequestedPredecessorRange(t *testing.T) {
	initTLCCheckerTest(t)
	worker := NewWorker(0)
	worker.SetTraceContext(t.TempDir(), "Spec")
	defer worker.CloseTrace()
	if err := worker.ensureTraceRAF(); err != nil {
		t.Fatal(err)
	}
	initial, middle, end := checkerTestState(1), checkerTestState(2), checkerTestState(3)
	if err := worker.WriteInitState(initial, initial.FingerPrint()); err != nil {
		t.Fatal(err)
	}
	if err := worker.WriteNextState(initial, middle.FingerPrint(), middle, nil); err != nil {
		t.Fatal(err)
	}
	if err := worker.WriteNextState(middle, end.FingerPrint(), end, nil); err != nil {
		t.Fatal(err)
	}
	trace := &ConcurrentTLCTrace{TLCTrace: NewTLCTrace(), Workers: []*Worker{worker}}
	lookups := 0
	trace.SetTool(&Tool{GetStateFunc: func(_ *Tool, fp uint64, prev ...any) (*TLCStateInfo, error) {
		lookups++
		if trace.mu.TryLock() {
			trace.mu.Unlock()
			t.Fatal("reconstruction did not retain monitor")
		}
		if fp == initial.FingerPrint() && len(prev) == 0 {
			return NewTLCStateInfo(initial), nil
		}
		if fp == middle.FingerPrint() && len(prev) == 1 && prev[0] == initial {
			return NewTLCStateInfo(middle), nil
		}
		t.Fatalf("unexpected reconstruction lookup %d/%v", fp, prev)
		return nil, nil
	}})
	got := trace.GetTraceFromState(end)
	if len(got) != 2 || got[0].State != initial || got[1].State != middle || lookups != 2 {
		t.Fatalf("full predecessor prefix: %v, lookups %d", got, lookups)
	}
	lookups = 0
	got = trace.GetTraceBetweenStates(initial, end)
	if len(got) != 2 || got[0].State != initial || got[1].State != middle || lookups != 1 {
		t.Fatalf("initial-anchored prefix: %v, lookups %d", got, lookups)
	}
	lookups = 0
	got = trace.GetTraceBetweenStates(middle, end)
	if len(got) != 1 || got[0].State != middle || lookups != 0 {
		t.Fatalf("middle-anchored prefix: %v, lookups %d", got, lookups)
	}
	if !trace.mu.TryLock() {
		t.Fatal("successful recovery retained monitor")
	}
	trace.mu.Unlock()
}
