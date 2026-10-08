package tlc

import "testing"

// No original method directly changes these owners after worker construction.
func TestWorkerSuccessorUsesCapturedFingerprintSet(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, missing := range []bool{false, true} {
		name := "replacement"
		if missing {
			name = "missing-replacement"
		}
		t.Run(name, func(t *testing.T) {
			original, replacement := NewMemFPSet(), NewMemFPSet()
			checker := NewModelChecker(NewTool(), t.TempDir(), false, WithModelCheckerFPSet(original))
			defer checker.ConcurrentTrace.Close()
			checker.FPSet = replacement
			if missing {
				checker.FPSet = nil
			}
			parent, successor := checkerTestState(1), checkerTestState(2)
			parent.UID = 37
			_, err := checker.Workers[0].AddNextElement(parent, &Action{Name: "Next"}, successor)
			if err != nil || original.Size() != 1 || replacement.Size() != 0 || checker.StateQueue.Size() != 1 {
				t.Fatalf("fingerprint owner changed: err %v, original %d, replacement %d, queue %d", err, original.Size(), replacement.Size(), checker.StateQueue.Size())
			}
		})
	}
}

func TestWorkerSuccessorUsesCapturedStateWriter(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, constrained := range []bool{false, true} {
		for _, missing := range []bool{false, true} {
			name := "successor"
			if constrained {
				name = "unsatisfied"
			}
			if missing {
				name += "/missing-replacement"
			} else {
				name += "/replacement"
			}
			t.Run(name, func(t *testing.T) {
				originalCalls, replacementCalls := 0, 0
				original := &StateWriter{Constrained: constrained, WriteTransitionFunc: func(_, _ *TLCStateMut, status StateVisitStatus, _ *Action, _ ...SemanticNode) error {
					originalCalls++
					want := StateVisitUnseen
					if constrained {
						want = StateVisitNotInModel
					}
					if status != want {
						t.Fatalf("status %v, want %v", status, want)
					}
					return nil
				}}
				replacement := &StateWriter{Constrained: constrained, WriteTransitionFunc: func(_, _ *TLCStateMut, _ StateVisitStatus, _ *Action, _ ...SemanticNode) error {
					replacementCalls++
					return nil
				}}
				checker := NewModelChecker(NewTool(), t.TempDir(), false, WithModelCheckerStateWriter(original))
				defer checker.ConcurrentTrace.Close()
				checker.AllStateWriter = replacement
				if missing {
					checker.AllStateWriter = nil
				}
				parent, successor := checkerTestState(1), checkerTestState(2)
				parent.UID = 37
				worker := checker.Workers[0]
				if constrained {
					if got := worker.AddUnsatisfiedNextState(parent, &Action{Name: "Next"}, successor, nil, nil); got != successor {
						t.Fatal("successor changed")
					}
				} else if _, err := worker.AddNextElement(parent, &Action{Name: "Next"}, successor); err != nil {
					t.Fatal(err)
				}
				if originalCalls != 1 || replacementCalls != 0 {
					t.Fatalf("writer owner changed: original %d, replacement %d", originalCalls, replacementCalls)
				}
			})
		}
	}
}

func TestWorkerConstrainedReasonsUseCapturedStateWriter(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	tool := NewTool()
	tool.ModelConstraints = []SemanticNode{"state constraint"}
	tool.ActionConstraints = []SemanticNode{"action constraint"}
	tool.IsInModelFunc = func(*Tool, *TLCStateMut) (bool, error) { return false, nil }
	tool.IsInModelForConstraintFunc = func(*Tool, SemanticNode, *TLCStateMut) (bool, error) { return false, nil }
	tool.IsInActionsForConstraintFn = func(*Tool, SemanticNode, *TLCStateMut, *TLCStateMut) (bool, error) { return false, nil }
	var reasons []SemanticNode
	original := &StateWriter{Constrained: true, WriteTransitionFunc: func(_, _ *TLCStateMut, status StateVisitStatus, _ *Action, reason ...SemanticNode) error {
		if status != StateVisitNotInModel {
			t.Fatalf("status %v", status)
		}
		reasons = append(reasons, reason...)
		return nil
	}}
	checker := NewModelChecker(tool, t.TempDir(), false, WithModelCheckerStateWriter(original))
	defer checker.ConcurrentTrace.Close()
	checker.AllStateWriter = NewNoopStateWriter()
	parent, successor := checkerTestState(1), checkerTestState(2)
	parent.UID = 37
	if _, err := checker.Workers[0].AddNextElement(parent, &Action{Name: "Next"}, successor); err != nil {
		t.Fatal(err)
	}
	if len(reasons) != 2 || reasons[0] != "state constraint" || reasons[1] != "action constraint" || checker.FPSet.Size() != 0 || checker.StateQueue.Size() != 0 {
		t.Fatalf("constrained publication changed: reasons %v, fingerprints %d, queue %d", reasons, checker.FPSet.Size(), checker.StateQueue.Size())
	}
}

func TestWorkerUnsatisfiedWriterFailurePropagates(t *testing.T) {
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	failure := NewRuntimeException("constrained writer failed")
	writer := &StateWriter{Constrained: true, WriteTransitionFunc: func(_, _ *TLCStateMut, _ StateVisitStatus, _ *Action, _ ...SemanticNode) error { return failure }}
	checker := NewModelChecker(NewTool(), t.TempDir(), false, WithModelCheckerStateWriter(writer))
	defer checker.ConcurrentTrace.Close()
	defer func() {
		if got := recover(); got != failure {
			t.Fatalf("writer failure = %v, want original %v", got, failure)
		}
	}()
	checker.Workers[0].AddUnsatisfiedNextState(checkerTestState(1), &Action{Name: "Next"}, checkerTestState(2), "constraint", nil)
}
