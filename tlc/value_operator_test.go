package tlc

import (
	"strings"
	"testing"
)

func TestCallableValueRequiresSuccessorStateLikeJava(t *testing.T) {
	callable := NewCallableValue("callableMethod", TLCLevelAction, nil, func(args []Value) (func() (any, error), error) {
		return func() (any, error) { return "ran", nil }, nil
	})

	value, err := callable.EvalWithTool(nil, nil, EmptyContext, nil, nil, EvalClear, DoNotRecordCostModel)
	if err == nil {
		t.Fatalf("EvalWithTool without successor state = (%v, nil), want Java method override error", value)
	}
	if !strings.Contains(err.Error(), "Attempted to apply the operator overridden by the Java method") ||
		!strings.Contains(err.Error(), "callableMethod") {
		t.Fatalf("missing successor state error = %q", err.Error())
	}

	successor := NewEmptyState()
	value, err = callable.EvalWithTool(nil, nil, EmptyContext, nil, successor, EvalClear, DoNotRecordCostModel)
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
