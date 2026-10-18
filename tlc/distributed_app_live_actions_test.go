package tlc

import (
	"reflect"
	"testing"
)

// Source TLCApp uses indexed loops over its public action arrays. No direct
// upstream test covers replacing those arrays during evaluator callbacks.
func TestDistributedAppGenerationReadsLiveActions(t *testing.T) {
	for _, mutation := range []string{"shrink", "replace", "grow"} {
		t.Run(mutation, func(t *testing.T) {
			first, old, newAction := &Action{Name: "first"}, &Action{Name: "old"}, &Action{Name: "new"}
			app := &TLCApp{Actions: []*Action{first, old}}
			if mutation == "grow" {
				app.Actions = []*Action{first}
			}
			var calls []*Action
			failure := NewRuntimeException("new action failed")
			app.Tool = &Tool{GetNextStatesFunc: func(_ *Tool, action *Action, _ *TLCStateMut) (*StateVec, error) {
				calls = append(calls, action)
				if action == first {
					if mutation == "shrink" {
						app.Actions = []*Action{}
					} else {
						app.Actions = []*Action{first, newAction}
					}
				}
				if action == newAction {
					return nil, failure
				}
				return NewStateVec(0), nil
			}}
			result, err := app.GetNextStates(nil)
			want := []*Action{first}
			if mutation == "shrink" {
				if err != nil || result == nil || result.Size() != 0 {
					t.Fatalf("shortened action list result %v/%v", result, err)
				}
			} else {
				want = append(want, newAction)
				if err != failure || result != nil {
					t.Fatalf("replacement action failure lost: %v/%v", result, err)
				}
			}
			if !reflect.DeepEqual(calls, want) {
				t.Fatalf("called %v, want %v", calls, want)
			}
		})
	}
}

func TestDistributedAppPropertiesReadLiveActions(t *testing.T) {
	for _, topic := range []string{"invariants", "implied-inits", "implied-actions"} {
		for _, mutation := range []string{"shrink", "replace", "grow"} {
			t.Run(topic+"/"+mutation, func(t *testing.T) {
				first, old, newAction := &Action{Name: "first"}, &Action{Name: "old"}, &Action{Name: "new"}
				successor, predecessor := &TLCStateMut{UID: 2}, &TLCStateMut{UID: 1}
				app := &TLCApp{}
				set := func(actions []*Action) {
					switch topic {
					case "invariants":
						app.Invariants = actions
					case "implied-inits":
						app.ImpliedInits = actions
					default:
						app.ImpliedActions = actions
					}
				}
				initial := []*Action{first, old}
				if mutation == "grow" {
					initial = []*Action{first}
				}
				set(initial)
				var calls []*Action
				valid := func(action *Action) (bool, error) {
					calls = append(calls, action)
					if action == first {
						if mutation == "shrink" {
							set([]*Action{})
						} else {
							set([]*Action{first, newAction})
						}
					}
					return action != newAction, nil
				}
				app.Tool = &Tool{InvariantNames: []string{"first", "current-name"}, ImpliedInitNames: []string{"first", "current-name"}, ImpliedActNames: []string{"first", "current-name"},
					IsValidStateFunc: func(_ *Tool, action *Action, got *TLCStateMut) (bool, error) {
						if got != successor {
							t.Fatal("property check changed successor")
						}
						return valid(action)
					}, IsValidTransitionFunc: func(_ *Tool, action *Action, p, s *TLCStateMut) (bool, error) {
						if p != predecessor || s != successor {
							t.Fatal("property check changed transition")
						}
						return valid(action)
					}}
				if topic == "implied-inits" {
					predecessor = nil
				}
				err := app.CheckState(predecessor, successor)
				want := []*Action{first}
				if mutation == "shrink" {
					if err != nil {
						t.Fatal(err)
					}
				} else {
					want = append(want, newAction)
					failure, ok := err.(*WorkerException)
					label := map[string]string{"invariants": "Invariant", "implied-inits": "Implied-init", "implied-actions": "Implied-action"}[topic]
					if !ok || failure.Msg != "Error: "+label+" current-name is violated." || failure.State1 != predecessor || failure.State2 != successor || failure.KeepCallStack {
						t.Fatalf("live property violation/context %v", err)
					}
				}
				if !reflect.DeepEqual(calls, want) {
					t.Fatalf("called %v, want %v", calls, want)
				}
			})
		}
	}
}
