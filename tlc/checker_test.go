package tlc

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestModelCheckerDoInitDeduplicatesAndSkipsDuplicateInvariantChecks(t *testing.T) {
	initTLCCheckerTest(t)
	state := checkerTestState(1)
	duplicate := checkerTestState(1)
	invariant := &Action{Name: "Inv"}
	invariantChecks := 0

	tool := NewTool()
	tool.InitStates = []*TLCStateMut{state, duplicate}
	tool.Invariants = []*Action{invariant}
	tool.InvariantNames = []string{"Inv"}
	tool.IsValidStateFunc = func(tl *Tool, action *Action, state *TLCStateMut) (bool, error) {
		if action == invariant {
			invariantChecks++
		}
		return true, nil
	}

	mc := NewModelChecker(tool, t.TempDir(), true)
	code, err := mc.DoInit(false)
	if err != nil {
		t.Fatalf("DoInit returned error: %v", err)
	}
	if code != NoError {
		t.Fatalf("DoInit code = %d, want %d", code, NoError)
	}
	if mc.GetInitialStatesGenerated() != 2 {
		t.Fatalf("initial states generated = %d, want 2", mc.GetInitialStatesGenerated())
	}
	if got := mc.GetDistinctStatesGenerated(); got != 1 {
		t.Fatalf("distinct states = %d, want 1", got)
	}
	if got := mc.GetStateQueueSize(); got != 1 {
		t.Fatalf("queued states = %d, want 1", got)
	}
	if invariantChecks != 1 {
		t.Fatalf("invariant checks = %d, want 1 for the unseen initial state only", invariantChecks)
	}
}

func TestModelCheckerDoInitForceChecksDuplicateInitialStates(t *testing.T) {
	initTLCCheckerTest(t)
	invariant := &Action{Name: "Inv"}
	invariantChecks := 0

	tool := NewTool()
	tool.InitStates = []*TLCStateMut{checkerTestState(1), checkerTestState(1)}
	tool.Invariants = []*Action{invariant}
	tool.InvariantNames = []string{"Inv"}
	tool.IsValidStateFunc = func(tl *Tool, action *Action, state *TLCStateMut) (bool, error) {
		if action == invariant {
			invariantChecks++
		}
		return true, nil
	}

	mc := NewModelChecker(tool, t.TempDir(), true)
	code, err := mc.DoInit(true)
	if err != nil {
		t.Fatalf("DoInit returned error: %v", err)
	}
	if code != NoError {
		t.Fatalf("DoInit code = %d, want %d", code, NoError)
	}
	if invariantChecks != 2 {
		t.Fatalf("invariant checks = %d, want 2 when forceChecks is true", invariantChecks)
	}
}

func TestModelCheckerDoNextEnqueuesOnlyUnseenInModelSuccessorsButChecksAllImpliedActions(t *testing.T) {
	initTLCCheckerTest(t)
	action := &Action{Name: "Next"}
	invariant := &Action{Name: "Inv"}
	implied := &Action{Name: "Imp"}
	cur := checkerTestState(1)
	seenSucc := checkerTestState(1)
	newSucc := checkerTestState(2)
	invariantChecks := 0
	impliedChecks := 0

	tool := NewTool()
	tool.Actions = []*Action{action}
	tool.Invariants = []*Action{invariant}
	tool.InvariantNames = []string{"Inv"}
	tool.ImpliedActions = []*Action{implied}
	tool.ImpliedActNames = []string{"Imp"}
	installTestNextStateGenerator(tool, func(tl *Tool, a *Action, state *TLCStateMut) (*StateVec, error) {
		return NewStateVecFrom([]*TLCStateMut{seenSucc, newSucc}), nil
	})
	tool.IsValidStateFunc = func(tl *Tool, a *Action, state *TLCStateMut) (bool, error) {
		if a == invariant {
			invariantChecks++
		}
		return true, nil
	}
	tool.IsValidTransitionFunc = func(tl *Tool, a *Action, s0 *TLCStateMut, s1 *TLCStateMut) (bool, error) {
		if a == implied {
			impliedChecks++
		}
		return true, nil
	}

	mc := NewModelChecker(tool, t.TempDir(), true)
	mc.FPSet.Put(cur.FingerPrint())
	stop, err := mc.DoNext(cur)
	if err != nil {
		t.Fatalf("DoNext returned error: %v", err)
	}
	if stop {
		t.Fatalf("DoNext stop = true, want false")
	}
	if mc.NextStatesGenerated != 2 {
		t.Fatalf("next states generated = %d, want 2", mc.NextStatesGenerated)
	}
	if got := mc.GetDistinctStatesGenerated(); got != 2 {
		t.Fatalf("distinct states = %d, want 2 including the seeded current state and one new successor", got)
	}
	if got := mc.GetStateQueueSize(); got != 1 {
		t.Fatalf("queued states = %d, want 1 new successor", got)
	}
	if invariantChecks != 1 {
		t.Fatalf("invariant checks = %d, want 1 for the unseen successor only", invariantChecks)
	}
	if impliedChecks != 2 {
		t.Fatalf("implied-action checks = %d, want 2 for seen and unseen successors", impliedChecks)
	}
	if records := mc.Trace.Records(); len(records) != 1 {
		t.Fatalf("trace records = %d, want 1 for the unseen successor", len(records))
	}
}

func TestModelCheckerDoNextReportsDeadlockWhenNoActionProducesSuccessors(t *testing.T) {
	initTLCCheckerTest(t)
	action := &Action{Name: "Next"}
	cur := checkerTestState(1)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)

	tool := NewTool()
	tool.Actions = []*Action{action}
	installTestNextStateGenerator(tool, func(tl *Tool, a *Action, state *TLCStateMut) (*StateVec, error) {
		return NewStateVec(0), nil
	})

	mc := NewModelChecker(tool, t.TempDir(), true)
	stop, err := mc.DoNext(cur)
	if err != nil {
		t.Fatalf("DoNext returned error: %v", err)
	}
	if !stop {
		t.Fatalf("DoNext stop = false, want true")
	}
	if mc.ErrorCode != ECTLCDeadlockReached {
		t.Fatalf("error code = %d, want %d", mc.ErrorCode, ECTLCDeadlockReached)
	}
	if mc.ErrState != cur {
		t.Fatalf("error state = %p, want current state %p", mc.ErrState, cur)
	}
	if !recorder.Recorded(ECTLCDeadlockReached) {
		t.Fatalf("deadlock error was not recorded")
	}
}

func TestModelCheckerCleanupPreservesFailureArtifactsLikeJava(t *testing.T) {
	tlcSetSystemProperty(modelCheckerVetoProperty, "false")
	t.Setenv("TLAGO_MODEL_CHECKER_VETO_CLEANUP", "false")

	failedMetadir := t.TempDir()
	failedArtifact := filepath.Join(failedMetadir, "trace.chkpt")
	if err := os.WriteFile(failedArtifact, []byte("keep me"), 0o644); err != nil {
		t.Fatalf("WriteFile failed artifact: %v", err)
	}
	failedFPSet := &recordingFPSet{NoopFPSet: NewNoopFPSet(nil)}
	failed := &ModelChecker{
		AbstractChecker: &AbstractChecker{Metadir: failedMetadir},
		CleanupEnabled:  true,
		FPSet:           failedFPSet,
	}
	if err := failed.Cleanup(false, true); err != nil {
		t.Fatalf("failure Cleanup returned error: %v", err)
	}
	if !failedFPSet.closed {
		t.Fatalf("failure cleanup did not close FPSet; Java closes resources before preserving artifacts")
	}
	if _, err := os.Stat(failedArtifact); err != nil {
		t.Fatalf("failure cleanup removed artifact; Java non-recursive delete preserves it: %v", err)
	}

	successMetadir := t.TempDir()
	successArtifact := filepath.Join(successMetadir, "trace.chkpt")
	if err := os.WriteFile(successArtifact, []byte("remove me"), 0o644); err != nil {
		t.Fatalf("WriteFile success artifact: %v", err)
	}
	success := &ModelChecker{
		AbstractChecker: &AbstractChecker{Metadir: successMetadir},
		CleanupEnabled:  true,
	}
	if err := success.Cleanup(true, true); err != nil {
		t.Fatalf("success Cleanup returned error: %v", err)
	}
	if _, err := os.Stat(successMetadir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("success cleanup metadir exists/error = %v, want deleted like Java recursive delete", err)
	}
}

type recordingFPSet struct {
	*NoopFPSet
	closed bool
}

func (s *recordingFPSet) Close() {
	s.closed = true
}

func initTLCCheckerTest(t *testing.T) {
	t.Helper()
	UniqueStringInitialize()
	SetStateVariables([]string{"x"})
	ClearMessageRecorders()
	t.Cleanup(ClearMessageRecorders)
}

func checkerTestState(x int32) *TLCStateMut {
	return NewEmptyState().Bind(UniqueStringOf("x"), NewIntValue(x))
}

func installTestNextStateGenerator(tool *Tool, generator func(*Tool, *Action, *TLCStateMut) (*StateVec, error)) {
	tool.GetNextStatesFunc = generator
	tool.GetNextStatesForActionFunc = func(tl *Tool, functor *NextStateFunctor, state *TLCStateMut, action *Action) (bool, error) {
		next, err := generator(tl, action, state)
		if err != nil {
			return true, err
		}
		for i := 0; next != nil && i < next.Size(); i++ {
			if _, err := functor.AddNextElement(state, action, next.At(i)); err != nil {
				return true, err
			}
			if functor.ShouldHalt() {
				return true, nil
			}
		}
		return false, nil
	}
}
