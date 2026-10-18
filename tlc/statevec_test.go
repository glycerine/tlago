package tlc

import (
	"math"
	"testing"
)

func TestStateVecSourceConstructorsPreserveActiveReferences(t *testing.T) {
	initTLCCheckerTest(t)
	state := checkerTestState(1)
	single := NewStateVecWithState(state)
	if single.Size() != 1 || single.ElementAt(0) != state || cap(single.states) != 1 {
		t.Fatalf("single-state constructor = size %d cap %d state %p", single.Size(), cap(single.states), single.At(0))
	}
	nullSingle := NewStateVecWithState(nil)
	if nullSingle.Size() != 1 || nullSingle.ElementAt(0) != nil {
		t.Fatalf("null single-state constructor = size %d value %v", nullSingle.Size(), nullSingle.ElementAt(0))
	}

	original := NewStateVec(4)
	original.Add(state)
	original.ReplaceAt(1, checkerTestState(2))
	copyVec := NewStateVecCopy(original)
	if copyVec.Size() != 1 || cap(copyVec.states) != 1 || copyVec.ElementAt(0) != state {
		t.Fatalf("copy constructor = size %d cap %d state %v", copyVec.Size(), cap(copyVec.states), copyVec.ElementAt(0))
	}
	original.ReplaceAt(0, checkerTestState(3))
	if copyVec.ElementAt(0) != state {
		t.Fatal("copy constructor retained the source backing array")
	}

	var nilVec *StateVec
	defer func() {
		if _, ok := recover().(*NullPointerException); !ok {
			t.Fatal("null copy source did not raise NullPointerException")
		}
	}()
	_ = NewStateVecCopy(nilVec)
}

func TestStateVecGrowCopiesOnlyActiveEntriesAndUsesSetBound(t *testing.T) {
	initTLCCheckerTest(t)
	Globals.Lock()
	oldBound := Globals.SetBound
	Globals.SetBound = 3
	Globals.Unlock()
	t.Cleanup(func() {
		Globals.Lock()
		Globals.SetBound = oldBound
		Globals.Unlock()
	})

	active := checkerTestState(1)
	overflow := NewStateVec(1)
	overflow.Grow(math.MaxInt32)
	if cap(overflow.states) != 2 {
		t.Fatalf("overflowing Java growth capacity = %d, want doubled capacity 2", cap(overflow.states))
	}

	vec := NewStateVec(2)
	vec.Add(active)
	vec.ReplaceAt(1, checkerTestState(2))
	vec.Grow(1)
	if vec.Size() != 1 || cap(vec.states) != 3 || vec.ElementAt(0) != active || vec.ElementAt(1) != nil {
		t.Fatalf("grown vector = size %d cap %d active %v inactive %v", vec.Size(), cap(vec.states), vec.ElementAt(0), vec.ElementAt(1))
	}

	defer func() {
		failure := recover()
		runtimeFailure, ok := failure.(*TLCError)
		if !ok || runtimeFailure.Code != ECTLCTooManyPossibleStates {
			t.Fatalf("bound failure = %T: %v, want TLCRuntimeException(%d)", failure, failure, ECTLCTooManyPossibleStates)
		}
	}()
	vec.Grow(1)
}

func TestStateVecGeneralStateAccessorsAndSetConstructor(t *testing.T) {
	first := newJavaDummyTLCState(11)
	last := newJavaDummyTLCState(13)
	vector := NewStateVecFromStates([]TLCState{first, last})
	if vector.FirstState() != first || vector.LastState() != last {
		t.Fatalf("general first/last = %v/%v, want original states", vector.FirstState(), vector.LastState())
	}
	list := vector.ToList()
	if len(list) != 2 || list[0] != first || list[1] != last {
		t.Fatalf("general list = %#v, want original states", list)
	}
	vector.ReplaceAt(0, last)
	if list[0] != first {
		t.Fatal("toList retained the vector backing array")
	}

	set := NewSetOfStatesFromVec(NewStateVecFromStates([]TLCState{first, last}))
	if set.Size() != 2 || !set.Put(first) || !set.Put(last) {
		t.Fatalf("StateVec set constructor lost general states: size %d", set.Size())
	}
}

func TestStateVecAddNextElementStampsPredecessorAndAction(t *testing.T) {
	initTLCCheckerTest(t)
	SetTLCStateTool(NewTool().SetMode(ModeSimulation))
	t.Cleanup(func() { SetTLCStateTool(nil) })
	pred := checkerTestState(1)
	succ := checkerTestState(2)
	action := &Action{Name: "Next"}

	vec := NewStateVec(0)
	ret, err := vec.AddNextElement(pred, action, succ)
	if err != nil {
		t.Fatalf("AddNextElement returned error: %v", err)
	}
	if ret != vec {
		t.Fatalf("AddNextElement return = %p, want vector %p", ret, vec)
	}
	if vec.Size() != 1 || vec.At(0) != succ {
		t.Fatalf("vector contents after AddNextElement = %v", vec)
	}
	if succ.Predecessor() != pred {
		t.Fatalf("successor predecessor = %p, want %p", succ.Predecessor(), pred)
	}
	if succ.GetAction() != action {
		t.Fatalf("successor action = %p, want %p", succ.GetAction(), action)
	}
	if succ.Level() != pred.Level()+1 {
		t.Fatalf("successor level = %d, want %d", succ.Level(), pred.Level()+1)
	}
}

func TestStateVecAddElementsUsesLargerVectorAsResult(t *testing.T) {
	initTLCCheckerTest(t)
	a := checkerTestState(1)
	b := checkerTestState(2)
	c := checkerTestState(3)

	small := NewStateVecFrom([]*TLCStateMut{a})
	large := NewStateVecFrom([]*TLCStateMut{b, c})
	result := small.AddElements(large)

	if result != large {
		t.Fatalf("AddElements returned %p, want larger vector %p", result, large)
	}
	if result.Size() != 3 {
		t.Fatalf("result size = %d, want 3", result.Size())
	}
	if result.At(0) != b || result.At(1) != c || result.At(2) != a {
		t.Fatalf("result order = %v, want larger vector contents followed by smaller contents", result)
	}
}

func TestStateVecAddElementsAppendsOtherWhenReceiverIsLarger(t *testing.T) {
	initTLCCheckerTest(t)
	a := checkerTestState(1)
	b := checkerTestState(2)
	c := checkerTestState(3)

	large := NewStateVecFrom([]*TLCStateMut{a, b})
	small := NewStateVecFrom([]*TLCStateMut{c})
	result := large.AddElements(small)

	if result != large {
		t.Fatalf("AddElements returned %p, want receiver %p", result, large)
	}
	if result.Size() != 3 {
		t.Fatalf("result size = %d, want 3", result.Size())
	}
	if result.At(0) != a || result.At(1) != b || result.At(2) != c {
		t.Fatalf("result order = %v, want receiver contents followed by other contents", result)
	}
}

func TestStateVecAddRespectsConfiguredSetBound(t *testing.T) {
	initTLCCheckerTest(t)
	Globals.Lock()
	oldBound := Globals.SetBound
	Globals.SetBound = 2
	Globals.Unlock()
	t.Cleanup(func() {
		Globals.Lock()
		Globals.SetBound = oldBound
		Globals.Unlock()
	})

	vec := NewStateVec(0)
	vec.Add(checkerTestState(1))
	vec.Add(checkerTestState(2))
	expectPanic(t, func() { vec.Add(checkerTestState(3)) })
}

func TestStateVecAddElementsRespectsConfiguredSetBound(t *testing.T) {
	initTLCCheckerTest(t)
	Globals.Lock()
	oldBound := Globals.SetBound
	Globals.SetBound = 2
	Globals.Unlock()
	t.Cleanup(func() {
		Globals.Lock()
		Globals.SetBound = oldBound
		Globals.Unlock()
	})

	left := NewStateVecFrom([]*TLCStateMut{checkerTestState(1)})
	right := NewStateVecFrom([]*TLCStateMut{checkerTestState(2), checkerTestState(3)})
	expectPanic(t, func() { left.AddElements(right) })
}
