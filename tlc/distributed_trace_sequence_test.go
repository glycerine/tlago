package tlc

import "testing"

// No original method directly covers the random-generator lifetime or missing
// initial result in TLCTrace's fingerprint-sequence reconstruction helper.
func TestDistributedTraceSequenceRecovery(t *testing.T) {
	initTLCCheckerTest(t)
	for _, scenario := range []string{"empty", "success", "provided-initial", "initial-error", "successor-error", "missing-initial", "missing-initial-with-successor"} {
		t.Run(scenario, func(t *testing.T) {
			prior := NewJavaRandom(123)
			prior.enumerableKind = randomEnumerableDefault
			old := SetRandomEnumerableGenerator(prior)
			defer SetRandomEnumerableGenerator(old)
			initial, successor := NewTLCStateInfo(checkerTestState(1)), NewTLCStateInfo(checkerTestState(2))
			originalFP := uint64(71)
			initial.FP, successor.FP = &originalFP, &originalFP
			fps := []uint64{23, 17}
			var supplied *TLCStateInfo
			failure := NewIOException("sequence evaluation failed")
			calls := 0
			tool := &Tool{GetStateFunc: func(_ *Tool, fp uint64, prev ...any) (*TLCStateInfo, error) {
				calls++
				if RandomEnumerableGenerator() == prior {
					t.Fatal("reconstruction did not reset random generator")
				}
				if len(prev) == 0 {
					if fp != 17 {
						t.Fatalf("initial fingerprint %d", fp)
					}
					if scenario == "initial-error" {
						return nil, failure
					}
					if scenario == "missing-initial" || scenario == "missing-initial-with-successor" {
						return nil, nil
					}
					return initial, nil
				}
				if len(prev) != 1 || prev[0] != initial.State || fp != 23 {
					t.Fatalf("successor arguments %d/%v", fp, prev)
				}
				if scenario == "successor-error" {
					return nil, failure
				}
				return successor, nil
			}}
			switch scenario {
			case "empty":
				fps = nil
			case "provided-initial":
				supplied = initial
			case "missing-initial":
				fps = []uint64{17}
			}
			trace := NewTLCTrace()
			trace.Tool = tool
			var got []*TLCStateInfo
			err := invokeDistributedServerOperation(func() error { var e error; got, e = trace.recoverTraceFromFPs(supplied, fps); return e })
			switch scenario {
			case "initial-error", "successor-error":
				if err != failure || RandomEnumerableGenerator() == prior {
					t.Fatalf("error identity/random lifetime changed: %T/%v", err, err)
				}
			case "missing-initial-with-successor":
				if _, ok := err.(*NullPointerException); !ok || calls != 1 || RandomEnumerableGenerator() == prior {
					t.Fatalf("missing initial dereference: %T/%v, calls %d", err, err, calls)
				}
			default:
				if err != nil || RandomEnumerableGenerator() != prior {
					t.Fatalf("normal recovery/restoration: %v", err)
				}
				if scenario == "empty" {
					if got == nil || len(got) != 0 || calls != 0 {
						t.Fatalf("empty result/calls: %v/%d", got, calls)
					}
				} else if scenario == "missing-initial" {
					if len(got) != 1 || got[0] != nil {
						t.Fatalf("source retains single missing initial: %v", got)
					}
				} else if len(got) != 2 || got[0] != initial || got[1] != successor || *initial.FP != 71 || *successor.FP != 71 {
					t.Fatal("recovery changed returned info identity or metadata")
				}
			}
		})
	}
}

func TestDistributedToolMissingStateOverloads(t *testing.T) {
	initTLCCheckerTest(t)
	tool := &Tool{}
	state := checkerTestState(1)
	for _, previous := range []any{nil, state, NewTLCStateInfo(state)} {
		var info *TLCStateInfo
		var err error
		if previous == nil {
			info, err = tool.GetState(23)
		} else {
			info, err = tool.GetState(23, previous)
		}
		if info != nil {
			t.Fatal("missing lookup fabricated state")
		}
		if _, wrapped := previous.(*TLCStateInfo); wrapped {
			failure, ok := err.(*EvalException)
			if !ok || failure.ErrorCode != ECTLCFailedToRecoverNext {
				t.Fatalf("info overload failure: %T/%v", err, err)
			}
		} else if err != nil {
			t.Fatalf("state overload missing result became error: %T/%v", err, err)
		}
	}
}
