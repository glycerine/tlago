package tlc

import "testing"

func TestTLCTraceAndToolStateRegistryRecoverBehaviorPath(t *testing.T) {
	initTLCCheckerTest(t)
	action := &Action{Name: "Next"}
	initState := checkerTestState(0)
	var succState *TLCStateMut

	tool := NewTool()
	tool.InitStates = []*TLCStateMut{initState}
	tool.Actions = []*Action{action}
	tool.GetNextStatesFunc = func(tl *Tool, a *Action, state *TLCStateMut) (*StateVec, error) {
		if checkerStateX(state) != 0 {
			return NewStateVec(0), nil
		}
		succState = checkerTestState(1)
		return NewStateVecFrom([]*TLCStateMut{succState}), nil
	}

	mc := NewModelChecker(tool, t.TempDir(), false)
	code, err := mc.ModelCheck()
	if err != nil {
		t.Fatalf("ModelCheck returned error: %v", err)
	}
	if code != NoError {
		t.Fatalf("ModelCheck code = %d, want %d", code, NoError)
	}
	if succState == nil {
		t.Fatalf("successor state was not generated")
	}

	initFP := initState.FingerPrint()
	initInfo, err := tool.GetState(initFP)
	if err != nil {
		t.Fatalf("GetState(init) returned error: %v", err)
	}
	if initInfo.State != initState {
		t.Fatalf("GetState(init) state = %p, want %p", initInfo.State, initState)
	}

	succFP := succState.FingerPrint()
	succInfo, err := tool.GetState(succFP, initInfo)
	if err != nil {
		t.Fatalf("GetState(successor) returned error: %v", err)
	}
	if succInfo.State != succState {
		t.Fatalf("GetState(successor) state = %p, want %p", succInfo.State, succState)
	}
	if succInfo.State.Predecessor() != initState {
		t.Fatalf("successor predecessor = %p, want %p", succInfo.State.Predecessor(), initState)
	}

	records := mc.Trace.Records()
	if len(records) != 2 {
		t.Fatalf("trace records = %d, want 2", len(records))
	}
	record, ok := mc.Trace.RecordFor(succState)
	if !ok {
		t.Fatalf("successor trace record not found")
	}
	if record.FP != succFP {
		t.Fatalf("successor trace fp = %d, want %d", record.FP, succFP)
	}
	if record.PreviousUID != initState.UID {
		t.Fatalf("successor predecessor uid = %d, want %d", record.PreviousUID, initState.UID)
	}

	trace := mc.Trace.GetTrace(succState)
	if len(trace) != 2 {
		t.Fatalf("trace length = %d, want 2", len(trace))
	}
	if checkerStateX(trace[0].State) != 0 || checkerStateX(trace[1].State) != 1 {
		t.Fatalf("trace states = [%d, %d], want [0, 1]", checkerStateX(trace[0].State), checkerStateX(trace[1].State))
	}
	between := mc.Trace.GetTraceBetween(initState, succState)
	if len(between) != 2 {
		t.Fatalf("between trace length = %d, want 2", len(between))
	}
}
