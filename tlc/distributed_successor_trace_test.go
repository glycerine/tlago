package tlc

import "testing"

func TestDistributedAppRetainsToolSuccessorMetadata(t *testing.T) {
	predecessor := &TLCStateMut{UID: 23, level: 4}
	successor := &TLCStateMut{UID: TLCStateInitUID, level: 5}
	action := &Action{}
	tool := &Tool{
		GetNextStatesFunc: func(_ *Tool, gotAction *Action, gotState *TLCStateMut) (*StateVec, error) {
			if gotAction != action || gotState != predecessor {
				t.Fatal("application changed tool arguments")
			}
			states := NewStateVec(1)
			states.Add(successor)
			return states, nil
		},
		IsGoodStateFunc: func(_ *Tool, state *TLCStateMut) bool { return state == successor },
	}
	app := &TLCApp{Tool: tool, Actions: []*Action{action}}
	states, err := app.GetNextStates(predecessor)
	if err != nil || states.Size() != 1 || states.At(0) != successor {
		t.Fatalf("successor result %v, error %v", states, err)
	}
	if successor.TracePredecessor() != nil || successor.GetAction() != nil || successor.Level() != 5 || successor.UID != TLCStateInitUID {
		t.Fatal("TLCApp injected metadata absent in Java TLCApp.getNextStates")
	}
	if _, err := EncodeDistributedStates([]*TLCStateMut{successor}); err != nil {
		t.Fatalf("ordinary tool successor cannot cross the native boundary: %v", err)
	}
}

func TestDistributedPublicationWritesIncomingPredecessorUID(t *testing.T) {
	trace := NewTLCTrace()
	manager := NewDistributedFPSetManagerFromFPSet(NewMemFPSet())
	thread := &TLCServerThread{Server: &TLCServer{Trace: trace, FPSetManager: manager}}
	state := &TLCStateMut{UID: 314, level: 6}
	states := NewStateVec(1)
	states.Add(state)
	fps := NewLongVecFrom([]int64{0x1234})
	thread.publishBlock(nil, []*StateVec{states}, []*LongVec{fps})
	records := trace.Records()
	if len(records) != 1 || records[0].PreviousUID != 314 || records[0].FP != 0x1234 {
		t.Fatalf("trace lost incoming predecessor location: %#v", records)
	}
	if state.UID != 0 || state.Level() != 6 || state.TracePredecessor() != nil || state.GetAction() != nil {
		t.Fatal("publication must replace UID without changing source state metadata")
	}
	if _, err := EncodeDistributedStates([]*TLCStateMut{state}); err != nil {
		t.Fatalf("published state cannot return to a worker: %v", err)
	}
}
