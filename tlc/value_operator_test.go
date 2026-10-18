package tlc

import (
	"strings"
	"testing"
)

func TestOpLambdaNullDefinitionMatchesJavaFailure(t *testing.T) {
	const message = `Cannot invoke "tla2sany.semantic.OpDefNode.getArity()" because "this.opDef" is null`
	unsourced := NewOpLambdaValue(nil, nil, nil, nil, nil, EvalClear)
	_, err := unsourced.Eval([]Value{}, EvalClear)
	if failure, ok := err.(*NullPointerException); !ok || failure.Error() != message {
		t.Fatalf("unsourced failure = %T/%v, want NullPointerException %q", err, err, message)
	}

	source := NewValueNode(NewIntValue(1))
	sourced := NewOpLambdaValue(nil, nil, nil, nil, nil, EvalClear)
	sourced.SetSource(source)
	_, err = sourced.Eval([]Value{}, EvalClear)
	fingerprint, ok := err.(*FingerprintException)
	if !ok {
		t.Fatalf("sourced failure = %T/%v, want FingerprintException", err, err)
	}
	if cause, ok := fingerprint.Cause.(*NullPointerException); !ok || cause.Error() != message {
		t.Fatalf("sourced cause = %T/%v, want NullPointerException %q", fingerprint.Cause, fingerprint.Cause, message)
	}
}

func TestCallableValueRequiresSuccessorStateLikeJava(t *testing.T) {
	oldTool := stateTool
	SetTLCStateTool(NewTool().SetMode(ModeExecutor))
	t.Cleanup(func() { SetTLCStateTool(oldTool) })
	callable := NewCallableValue("callableMethod", TLCLevelAction, nil, func(args []Value) (func() (any, error), error) {
		return func() (any, error) { return "ran", nil }, nil
	})

	value, err := callable.EvalWithTool(nil, []SemanticNode{}, EmptyContext, nil, nil, EvalClear, DoNotRecordCostModel)
	if err == nil {
		t.Fatalf("EvalWithTool without successor state = (%v, nil), want Java method override error", value)
	}
	if !strings.Contains(err.Error(), "Attempted to apply the operator overridden by the Java method") ||
		!strings.Contains(err.Error(), "callableMethod") {
		t.Fatalf("missing successor state error = %q", err.Error())
	}

	successor := NewEmptyState()
	value, err = callable.EvalWithTool(nil, []SemanticNode{}, EmptyContext, nil, successor, EvalClear, DoNotRecordCostModel)
	if err != nil {
		t.Fatalf("EvalWithTool with successor state returned error: %v", err)
	}
	if value != BoolTrue {
		t.Fatalf("EvalWithTool value = %v, want TRUE", value)
	}
	result, err := successor.ExecCallable()
	if err != nil {
		t.Fatalf("ExecCallable returned error: %v", err)
	}
	if result != "ran" {
		t.Fatalf("ExecCallable result = %v, want ran", result)
	}
}

// The upstream state classes have no direct callable test. Ordinary TLCStateMut
// inherits no-op setCallable/execCallable; TLCStateMutExt stores and executes it.
func TestStateCallableRespectsModelCheckingMode(t *testing.T) {
	oldPolicy := statePreserveMetadata
	t.Cleanup(func() { statePreserveMetadata = oldPolicy })
	for _, extended := range []bool{false, true} {
		name := "model_checking"
		if extended {
			name = "extended"
		}
		t.Run(name, func(t *testing.T) {
			statePreserveMetadata = extended
			state := NewEmptyState()
			calls := 0
			callback := func() (any, error) { calls++; return "ran", nil }
			value, err := TLCExtTLCDefer([]*TLCStateMut{state}, callback)
			if err != nil || value != BoolTrue || calls != 0 {
				t.Fatalf("defer = %v/%v, calls=%d", value, err, calls)
			}
			result, err := state.ExecCallable()
			if err != nil {
				t.Fatal(err)
			}
			if extended {
				if calls != 1 || result != "ran" || state.callable == nil {
					t.Fatal("extended state did not retain and execute callable")
				}
				if _, err := EncodeDistributedStates([]*TLCStateMut{state}); err == nil {
					t.Fatal("executable callable was silently transferred")
				}
			} else {
				if calls != 0 || result != nil || state.callable != nil {
					t.Fatal("ordinary model-checking state retained or executed callable")
				}
				if _, err := EncodeDistributedStates([]*TLCStateMut{state}); err != nil {
					t.Fatalf("ignored callable prevented native transfer: %v", err)
				}
			}
			state.SetCallable(nil)
			if result, err := state.ExecCallable(); result != nil || err != nil {
				t.Fatalf("cleared callable = %v/%v", result, err)
			}
		})
	}
}
