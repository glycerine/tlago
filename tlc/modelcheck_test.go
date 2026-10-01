package tlc

import "testing"

func TestModelCheckerModelCheckExploresFiniteStateGraph(t *testing.T) {
	initTLCCheckerTest(t)
	action := &Action{Name: "Next"}
	postChecks := 0

	tool := NewTool()
	tool.InitStates = []*TLCStateMut{checkerTestState(0)}
	tool.Actions = []*Action{action}
	installTestNextStateGenerator(tool, func(tl *Tool, a *Action, state *TLCStateMut) (*StateVec, error) {
		switch checkerStateX(state) {
		case 0:
			return NewStateVecFrom([]*TLCStateMut{checkerTestState(1)}), nil
		default:
			return NewStateVec(0), nil
		}
	})
	tool.CheckPostConditionFunc = func(tl *Tool) int {
		postChecks++
		return NoError
	}

	mc := NewModelChecker(tool, t.TempDir(), false)
	code, err := mc.ModelCheck()
	if err != nil {
		t.Fatalf("ModelCheck returned error: %v", err)
	}
	if code != NoError {
		t.Fatalf("ModelCheck code = %d, want %d", code, NoError)
	}
	if got := mc.GetDistinctStatesGenerated(); got != 2 {
		t.Fatalf("distinct states = %d, want 2", got)
	}
	if got := mc.GetStatesGenerated(); got != 2 {
		t.Fatalf("states generated = %d, want 2", got)
	}
	if got := mc.GetStateQueueSize(); got != 0 {
		t.Fatalf("queued states = %d, want 0 after exploration", got)
	}
	if postChecks != 1 {
		t.Fatalf("postcondition checks = %d, want 1", postChecks)
	}
}

func TestModelCheckerModelCheckReportsDeadlockFromWorkerLoop(t *testing.T) {
	initTLCCheckerTest(t)
	action := &Action{Name: "Next"}

	tool := NewTool()
	tool.InitStates = []*TLCStateMut{checkerTestState(0)}
	tool.Actions = []*Action{action}
	installTestNextStateGenerator(tool, func(tl *Tool, a *Action, state *TLCStateMut) (*StateVec, error) {
		return NewStateVec(0), nil
	})

	mc := NewModelChecker(tool, t.TempDir(), true)
	code, err := mc.ModelCheck()
	if err != nil {
		t.Fatalf("ModelCheck returned error: %v", err)
	}
	if code != ECTLCDeadlockReached {
		t.Fatalf("ModelCheck code = %d, want %d", code, ECTLCDeadlockReached)
	}
	if mc.ErrState == nil || checkerStateX(mc.ErrState) != 0 {
		t.Fatalf("deadlock state = %v, want x = 0", mc.ErrState)
	}
}

func TestModelCheckerHonorsConfigCheckDeadlockFalse(t *testing.T) {
	initTLCCheckerTest(t)
	action := &Action{Name: "Next"}

	cfg, err := ParseModelConfigSource("Spec.cfg", "CHECK_DEADLOCK FALSE\n")
	if err != nil {
		t.Fatalf("ParseModelConfigSource returned error: %v", err)
	}
	tool := NewToolWithModelConfig(cfg)
	tool.InitStates = []*TLCStateMut{checkerTestState(0)}
	tool.Actions = []*Action{action}
	installTestNextStateGenerator(tool, func(tl *Tool, a *Action, state *TLCStateMut) (*StateVec, error) {
		return NewStateVec(0), nil
	})

	mc := NewModelChecker(tool, t.TempDir(), true)
	code, err := mc.ModelCheck()
	if err != nil {
		t.Fatalf("ModelCheck returned error: %v", err)
	}
	if code != NoError {
		t.Fatalf("ModelCheck code = %d, want %d when CHECK_DEADLOCK is FALSE", code, NoError)
	}
	if mc.CheckDeadlock {
		t.Fatalf("checker deadlock flag = true, want false from config")
	}
}

func TestModelCheckerKeepsDeadlockDisabledWhenCallerDisablesIt(t *testing.T) {
	initTLCCheckerTest(t)
	action := &Action{Name: "Next"}

	cfg, err := ParseModelConfigSource("Spec.cfg", "CHECK_DEADLOCK TRUE\n")
	if err != nil {
		t.Fatalf("ParseModelConfigSource returned error: %v", err)
	}
	tool := NewToolWithModelConfig(cfg)
	tool.InitStates = []*TLCStateMut{checkerTestState(0)}
	tool.Actions = []*Action{action}
	installTestNextStateGenerator(tool, func(tl *Tool, a *Action, state *TLCStateMut) (*StateVec, error) {
		return NewStateVec(0), nil
	})

	mc := NewModelChecker(tool, t.TempDir(), false)
	code, err := mc.ModelCheck()
	if err != nil {
		t.Fatalf("ModelCheck returned error: %v", err)
	}
	if code != NoError {
		t.Fatalf("ModelCheck code = %d, want %d when caller disables deadlock", code, NoError)
	}
	if mc.CheckDeadlock {
		t.Fatalf("checker deadlock flag = true, want false from caller")
	}
}

func TestModelCheckerModelCheckRejectsStatesWithoutNextAction(t *testing.T) {
	initTLCCheckerTest(t)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)

	tool := NewTool()
	tool.InitStates = []*TLCStateMut{checkerTestState(0)}

	mc := NewModelChecker(tool, t.TempDir(), true)
	code, err := mc.ModelCheck()
	if err != nil {
		t.Fatalf("ModelCheck returned error: %v", err)
	}
	if code != ECTLCStatesAndNoNextAction {
		t.Fatalf("ModelCheck code = %d, want %d", code, ECTLCStatesAndNoNextAction)
	}
	if !recorder.Recorded(ECTLCStatesAndNoNextAction) {
		t.Fatalf("states-without-next-action error was not recorded")
	}
}

func checkerStateX(state *TLCStateMut) int32 {
	if state == nil {
		return 0
	}
	value, _ := state.Lookup(UniqueStringOf("x")).(*IntValue)
	if value == nil {
		return 0
	}
	return value.Val
}
