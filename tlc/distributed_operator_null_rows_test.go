package tlc

import "testing"

// The source OpRcdValue dereferences argument arrays. A nil row is not a
// zero-argument row even though Go len/range treat the two alike. No upstream
// method directly tests this network/materialization boundary.
func TestDistributedOperatorNullRowsFailEvaluation(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		domain [][]Value
		args   []Value
	}{
		{"nil_arguments", [][]Value{{}}, nil},
		{"nil_row", [][]Value{nil}, []Value{}},
		{"nil_arguments_empty_operator", [][]Value{}, nil},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			operator := NewOpRcdValueFrom(scenario.domain, []Value{NewIntValue(7)})
			if len(scenario.domain) == 0 {
				operator.Values = []Value{}
			}
			got := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{operator}}})[0].values[0].(*OpRcdValue)
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("null argument array did not fail as in the source: %T/%v", failure, failure)
				}
			}()
			value, err := got.Eval(scenario.args, 0)
			t.Fatalf("null operator input was accepted or recategorized: %v/%v", value, err)
		})
	}
	operator := NewOpRcdValueFrom([][]Value{{}}, []Value{NewIntValue(7)})
	value, err := operator.Eval([]Value{}, 0)
	if err != nil || value != operator.Values[0] {
		t.Fatalf("real zero-argument row no longer applies: %v/%v", value, err)
	}
}

func TestDistributedOperatorNullEntriesFailInitialization(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		domain [][]Value
		values []Value
	}{
		{"nil_row", [][]Value{nil}, []Value{NewIntValue(7)}},
		{"nil_argument", [][]Value{{nil}}, []Value{NewIntValue(7)}},
		{"nil_result", [][]Value{{}}, []Value{nil}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			operator := NewOpRcdValueFrom(scenario.domain, scenario.values)
			got := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{operator}}})[0].values[0].(*OpRcdValue)
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("null operator entry did not fail as in the source: %T/%v", failure, failure)
				}
			}()
			InitializeValue(got)
			t.Fatal("null operator entry was silently initialized")
		})
	}
	operator := NewOpRcdValueFrom([][]Value{{}}, []Value{NewIntValue(7)})
	if InitializeValue(operator) != operator {
		t.Fatal("real empty argument row no longer initializes")
	}
}

func TestDistributedWorkerNullOperatorRowFailure(t *testing.T) {
	for _, remote := range []bool{false, true} {
		name := "local"
		if remote {
			name = "tcp"
		}
		t.Run(name, func(t *testing.T) {
			worker, _ := managerOwnershipWorker(t, 0, nil)
			worker.App.Tool.GetNextStatesFunc = func(_ *Tool, _ *Action, predecessor *TLCStateMut) (*StateVec, error) {
				operator := predecessor.values[0].(*OpRcdValue)
				_, err := operator.Eval([]Value{}, 0)
				return NewStateVec(0), err
			}
			predecessor := &TLCStateMut{UID: 31, level: 4, values: []Value{NewOpRcdValueFrom([][]Value{nil}, []Value{NewIntValue(7)})}}
			invoke := worker.GetNextStates
			if remote {
				_, client := startWorkerRPC(t, NewLocalWorkerEndpoint(worker))
				invoke = client.GetNextStates
			}
			result, err := invoke([]*TLCStateMut{predecessor})
			failure, ok := err.(*WorkerException)
			if result != nil || !ok || !isDistributedNullFailure(failure.Cause) || !failure.KeepCallStack || failure.State1 == nil || failure.State1.UID != 31 || failure.State2 != nil {
				t.Fatalf("operator failure lost worker evaluation context: %v/%v", result, err)
			}
			if !remote && failure.State1 != predecessor {
				t.Fatal("local operator failure replaced predecessor")
			}
			if worker.Computing.Load() || worker.OverallStatesComputed.Load() != 0 || worker.LastInvocation.Load() == 0 {
				t.Fatal("operator failure changed worker statistics/finally order")
			}
		})
	}
}
