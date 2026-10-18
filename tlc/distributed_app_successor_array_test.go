package tlc

import (
	"reflect"
	"testing"
)

// TLCApp has no direct upstream successor-array test. Its tool vector is a
// mutable accumulator, but its returned array is a separate fixed-size object.
func TestDistributedAppSuccessorArrayOwnershipAndOrder(t *testing.T) {
	a, b, c := &TLCStateMut{UID: 1}, &TLCStateMut{UID: 2}, &TLCStateMut{UID: 3}
	first, second := &Action{}, &Action{}
	small, large := NewStateVecFrom([]*TLCStateMut{a}), NewStateVecFrom([]*TLCStateMut{b, c})
	var checked []*TLCStateMut
	tool := &Tool{GetNextStatesFunc: func(_ *Tool, action *Action, _ *TLCStateMut) (*StateVec, error) {
		if action == first {
			return small, nil
		}
		return large, nil
	}, IsGoodStateFunc: func(_ *Tool, state *TLCStateMut) bool { checked = append(checked, state); return true }}
	app := &TLCApp{Tool: tool, Actions: []*Action{first, second}}
	got, err := app.GetNextStates(&TLCStateMut{})
	if err != nil {
		t.Fatal(err)
	}
	if got == large || !reflect.DeepEqual(checked, []*TLCStateMut{b, c, a}) || got.Size() != 3 || cap(got.states) != 3 {
		t.Fatal("application returned accumulator or lost source larger-vector order")
	}
	large.states[0] = a
	large.Clear()
	if got.Size() != 3 || got.At(0) != b || got.At(1) != c || got.At(2) != a {
		t.Fatal("tool-vector mutation changed returned successor array")
	}
	got.states[1] = a
	if small.At(0) != a || large.states[:cap(large.states)][1] != c {
		t.Fatal("returned array mutation changed tool vector")
	}
	b.UID = 9
	if got.At(0).UID != 9 {
		t.Fatal("application deep-copied shared state object")
	}
}

func TestDistributedAppSuccessorArrayValidationMutation(t *testing.T) {
	for _, mutation := range []string{"replace", "shrink", "grow"} {
		t.Run(mutation, func(t *testing.T) {
			a, b, c := &TLCStateMut{UID: 1}, &TLCStateMut{UID: 2}, &TLCStateMut{UID: 3}
			vector := NewStateVecFrom([]*TLCStateMut{a, b})
			var checked []*TLCStateMut
			tool := &Tool{GetNextStatesFunc: func(*Tool, *Action, *TLCStateMut) (*StateVec, error) { return vector, nil },
				IsGoodStateFunc: func(_ *Tool, state *TLCStateMut) bool {
					checked = append(checked, state)
					if len(checked) == 1 {
						switch mutation {
						case "replace":
							vector.states[0] = c
						case "shrink":
							vector.Clear()
							vector.Add(c)
						case "grow":
							vector.Add(c)
						}
					}
					return true
				}}
			app := &TLCApp{Tool: tool, Actions: []*Action{{}}}
			if mutation == "grow" {
				defer func() {
					failure := recover()
					if _, ok := failure.(*ArrayIndexOutOfBoundsException); !ok {
						t.Fatalf("fixed-array overflow %T/%v", failure, failure)
					}
					if !reflect.DeepEqual(checked, []*TLCStateMut{a, b, c}) || vector.Size() != 3 {
						t.Fatal("array overflow happened before validating added successor")
					}
				}()
				_, _ = app.GetNextStates(nil)
				return
			}
			got, err := app.GetNextStates(nil)
			if err != nil || got.Size() != 2 || got.At(0) != a {
				t.Fatalf("validated state/result length changed: %v/%v", got, err)
			}
			if mutation == "replace" {
				if got.At(1) != b || len(checked) != 2 {
					t.Fatal("replacement changed captured first state or later validation")
				}
			} else if got.At(1) != nil || len(checked) != 1 {
				t.Fatal("shrinking vector lost original array length/null tail")
			}
		})
	}
}
