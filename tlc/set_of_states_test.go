package tlc

import "testing"

func TestSetOfStatesSizeAndGrowth(t *testing.T) {
	initTLCCheckerTest(t)
	set := NewSetOfStates(16)
	if set.Capacity() != 16 || set.Size() != 0 {
		t.Fatalf("empty capacity/size = %d/%d, want 16/0", set.Capacity(), set.Size())
	}
	if duplicate := set.Put(checkerTestState(1)); duplicate {
		t.Fatalf("first Put reported duplicate")
	}
	if set.Capacity() != 16 || set.Size() != 1 {
		t.Fatalf("after put capacity/size = %d/%d, want 16/1", set.Capacity(), set.Size())
	}

	grown := NewSetOfStates(1)
	for i := int32(0); i < 32; i++ {
		grown.Put(checkerTestState(i))
	}
	if grown.Capacity() <= 32 {
		t.Fatalf("grown capacity = %d, want > 32", grown.Capacity())
	}
	if grown.Size() != 32 {
		t.Fatalf("grown size = %d, want 32", grown.Size())
	}
}

func TestSetOfStatesIteratesAllStatesAndResets(t *testing.T) {
	initTLCCheckerTest(t)
	set := NewSetOfStates(1)
	for i := int32(1); i <= 32; i++ {
		if duplicate := set.Put(checkerTestState(i)); duplicate {
			t.Fatalf("Put(%d) duplicate on first insert", i)
		}
	}

	var sum int32
	var predecessor *TLCStateMut
	for i := 0; i < set.Size(); i++ {
		state := set.Next()
		if state == predecessor {
			t.Fatalf("Next returned same state twice in a row")
		}
		sum += state.Lookup(UniqueStringOf("x")).(*IntValue).Val
		predecessor = state
	}
	if sum != (32/2)*(1+32) {
		t.Fatalf("iteration sum = %d, want %d", sum, (32/2)*(1+32))
	}

	set.ResetNext()
	if set.Next() == nil {
		t.Fatalf("Next after ResetNext returned nil")
	}
}

func TestSetOfStatesDuplicatesUseStateEqualityAfterFingerprintCollision(t *testing.T) {
	initTLCCheckerTest(t)
	set := NewSetOfStates(16)
	a := checkerTestState(1)
	b := checkerTestState(2)
	c := checkerTestState(1)
	const forcedFP uint64 = 42

	if duplicate := set.PutFP(forcedFP, a); duplicate {
		t.Fatalf("first state reported duplicate")
	}
	if duplicate := set.PutFP(forcedFP, b); duplicate {
		t.Fatalf("same fingerprint but unequal state reported duplicate")
	}
	if duplicate := set.PutFP(forcedFP, c); !duplicate {
		t.Fatalf("same fingerprint and equal state did not report duplicate")
	}
	if set.Size() != 2 {
		t.Fatalf("size = %d, want 2", set.Size())
	}
}

func TestSetOfStatesSubsetUsesActionIdentity(t *testing.T) {
	initTLCCheckerTest(t)
	SetTLCStateTool(NewTool().SetMode(ModeSimulation))
	t.Cleanup(func() { SetTLCStateTool(nil) })
	actionA := &Action{Name: "A"}
	actionB := &Action{Name: "A"}
	stateA := checkerTestState(1).SetAction(actionA)
	stateB := checkerTestState(2).SetAction(actionB)

	set := NewSetOfStates(1)
	set.Put(stateA)
	set.Put(stateB)
	subset := set.GetSubSet(actionA)

	if len(subset) != 1 || subset[0] != stateA {
		t.Fatalf("subset = %#v, want only stateA", subset)
	}
	if set.Next() == nil {
		t.Fatalf("GetSubSet did not reset iterator")
	}
}
