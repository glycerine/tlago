package tlc

import "testing"

func TestStateVecAddNextElementStampsPredecessorAndAction(t *testing.T) {
	initTLCCheckerTest(t)
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
