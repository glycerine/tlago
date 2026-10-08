package tlc

import "testing"

// No original method directly covers absent reconstruction-tool ownership.
func TestTraceRecoveryRequiresToolOnlyAtLookup(t *testing.T) {
	initTLCCheckerTest(t)
	for _, kind := range []string{"single", "concurrent"} {
		for _, scenario := range []string{"empty", "provided-initial", "initial-lookup", "successor-lookup"} {
			t.Run(kind+"/"+scenario, func(t *testing.T) {
				prior := NewJavaRandom(123)
				prior.enumerableKind = randomEnumerableDefault
				old := SetRandomEnumerableGenerator(prior)
				defer SetRandomEnumerableGenerator(old)
				initial := NewTLCStateInfo(checkerTestState(1))
				initial.State.UID, initial.State.WorkerID = 77, 7
				fps := []uint64{17}
				records := []ConcurrentTraceRecord{{Ptr: 11, Worker: 4}, {Ptr: 1, FP: 17}}
				var supplied *TLCStateInfo
				switch scenario {
				case "empty":
					fps = nil
					records = records[1:]
				case "provided-initial":
					supplied = initial
				case "successor-lookup":
					supplied = initial
					fps = []uint64{23, 17}
					records = append([]ConcurrentTraceRecord{{Ptr: 22, Worker: 5}}, records...)
				}
				var got []*TLCStateInfo
				err := invokeDistributedServerOperation(func() error {
					var e error
					if kind == "single" {
						got, e = NewTLCTrace().recoverTraceFromFPs(supplied, fps)
					} else {
						got, e = (&ConcurrentTLCTrace{}).recoverTraceFromRecords(supplied, records)
					}
					return e
				})
				if scenario == "initial-lookup" || scenario == "successor-lookup" {
					if _, ok := err.(*NullPointerException); !ok || RandomEnumerableGenerator() == prior {
						t.Fatalf("lookup tool failure/random lifetime: %T/%v", err, err)
					}
				} else {
					if err != nil || got == nil || RandomEnumerableGenerator() != prior {
						t.Fatalf("tool-free path result/random: %v/%v", got, err)
					}
					if scenario == "empty" && len(got) != 0 {
						t.Fatal("empty path regenerated state")
					}
					if scenario == "provided-initial" && (len(got) != 1 || got[0] != initial) {
						t.Fatal("provided initial required tool or changed result")
					}
				}
				if initial.State.UID != 77 || initial.State.WorkerID != 7 {
					t.Fatal("tool-free path changed provided metadata")
				}
			})
		}
	}
}

func TestTracePrintingMissingTool(t *testing.T) {
	initTLCCheckerTest(t)
	for _, scenario := range []string{"initial-only", "initial-transition", "noninitial"} {
		t.Run(scenario, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			recorder := &MemoryRecorder{}
			AddMessageRecorder(recorder)
			defer RemoveMessageRecorder(recorder)
			current, successor := checkerTestState(1), checkerTestState(2)
			if scenario == "initial-only" {
				successor = nil
			}
			if scenario == "noninitial" {
				current.level = 2
				successor = nil
			}
			err := invokeDistributedServerOperation(func() error { NewTLCTrace().printTraceWithPrefix(current, successor, nil); return nil })
			wantStates := 0
			if scenario == "initial-only" {
				wantStates = 1
				if err != nil {
					t.Fatalf("initial-only printing required tool: %v", err)
				}
			} else if _, ok := err.(*NullPointerException); !ok {
				t.Fatalf("missing print tool: %T/%v", err, err)
			}
			if len(recorder.Records(ECTLCBehaviorUpToThisPoint)) != 1 || len(recorder.Records(ECTLCStatePrint2)) != wantStates || recorder.Recorded(ECTLCBug) || recorder.Recorded(ECTLCFailedToRecoverInit) {
				t.Fatal("missing tool changed preceding output or entered missing-match exit")
			}
		})
	}
}

func TestDiskTraceMissingToolDoesNotSelectFallback(t *testing.T) {
	initTLCCheckerTest(t)
	initial := checkerTestState(1)
	trace := NewTLCTrace(t.TempDir(), "Spec")
	defer trace.Close()
	if _, err := trace.WriteStateRecord(nil, initial.FingerPrint(), initial); err != nil {
		t.Fatal(err)
	}
	err := invokeDistributedServerOperation(func() error { trace.GetTraceAt(initial.UID, true); return nil })
	if _, ok := err.(*NullPointerException); !ok {
		t.Fatalf("disk reconstruction substituted in-memory trace: %T/%v", err, err)
	}
	if got := trace.GetTraceAt(initial.UID, false); got == nil || len(got) != 0 {
		t.Fatal("empty disk prefix required tool or substituted fallback")
	}
}

func TestTracePrintingDoesNotReplaceMissingAlias(t *testing.T) {
	initTLCCheckerTest(t)
	captureFailoverToolIO(t, ToolIOTool)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)
	trace := NewTLCTrace()
	trace.Tool = &Tool{EvalAliasInfoFunc: func(*Tool, *TLCStateInfo, *TLCStateMut, func() []*TLCStateInfo) (*TLCStateInfo, error) {
		return nil, nil
	}}
	err := invokeDistributedServerOperation(func() error { trace.printTraceWithPrefix(checkerTestState(1), checkerTestState(2), nil); return nil })
	if _, ok := err.(*NullPointerException); !ok {
		t.Fatalf("missing alias result failure: %T/%v", err, err)
	}
	if len(recorder.Records(ECTLCBehaviorUpToThisPoint)) != 1 || recorder.Recorded(ECTLCStatePrint2) {
		t.Fatal("missing alias result printed an unaliased substitute")
	}
}
