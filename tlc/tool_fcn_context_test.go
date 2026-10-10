package tlc

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestToolFunctionContextUsesBranchSpecificDiagnostics(t *testing.T) {
	a, b := NewSymbolNode("a"), NewSymbolNode("b")
	for _, tc := range []struct {
		name   string
		params *FcnParams
		arg    Value
		want   func(*FcnLambdaValue) string
	}{
		{
			name:   "single domain",
			params: NewSingleFcnParam(a, NewIntervalValue(1, 2)),
			arg:    NewIntValue(3),
			want: func(fcn *FcnLambdaValue) string {
				return "In applying the function\n" + ValuesPPR(fcn) + ",\nthe first argument is:\n3which is not in its domain.\nF"
			},
		},
		{
			name:   "multiple non-tuple",
			params: NewFcnParams([][]*SymbolNode{{a, b}}, []bool{false}, []Value{NewIntervalValue(1, 2)}),
			arg:    NewIntValue(1),
			want: func(*FcnLambdaValue) string {
				return "Attempted to apply a function to an argument not in its domain.\nF"
			},
		},
		{
			name:   "multiple domain",
			params: NewFcnParams([][]*SymbolNode{{a, b}}, []bool{false}, []Value{NewIntervalValue(1, 2)}),
			arg:    NewTupleValue([]Value{NewIntValue(1), NewIntValue(3)}),
			want: func(fcn *FcnLambdaValue) string {
				return "In applying the function\n" + ValuesPPR(fcn) + ",\nthe argument number 2 is:\n3which is not in its domain.\nF"
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tool := NewTool()
			fcn := NewFcnLambdaValue(tc.params, BoolTrue, tool, EmptyContext, EmptyState, EmptyState, EvalClear)
			function := NewValueNode(fcn)
			function.Image = "F"
			expr := NewBuiltinOpApplNode(OpFA, function, tc.arg)
			want := tc.want(fcn)
			_, err := tool.getFcnContext(fcn, expr, EmptyContext, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel)
			if err == nil || err.Error() != want {
				t.Fatalf("function-context error = %v, want %q", err, want)
			}
		})
	}
}

func TestToolFunctionContextPreservesCapturedContextAndMixedBindings(t *testing.T) {
	a, b, c, d, e := NewSymbolNode("a"), NewSymbolNode("b"), NewSymbolNode("c"), NewSymbolNode("d"), NewSymbolNode("e")
	captured := NewSymbolNode("captured")
	inner := NewFcnRcdIntervalValue(NewIntervalValue(1, 2), []Value{NewIntValue(7), NewIntValue(8)})
	params := NewFcnParams([][]*SymbolNode{{a, b}, {c, d}, {e}}, []bool{false, true, false}, []Value{
		NewIntervalValue(1, 2), NewSetEnumValue([]Value{inner}, true), NewIntervalValue(9, 9),
	})
	tool := NewTool()
	fcn := NewFcnLambdaValue(params, BoolTrue, tool, EmptyContext.Cons(captured, NewIntValue(42)), EmptyState, EmptyState, EvalClear)
	// Java ignores extra arguments after all formals have been bound.
	arg := NewTupleValue([]Value{NewIntValue(1), NewIntValue(2), inner, NewIntValue(9), NewIntValue(99)})
	expr := NewBuiltinOpApplNode(OpFA, fcn, arg)
	ctx, err := tool.getFcnContext(fcn, expr, EmptyContext.Cons(captured, NewIntValue(0)), EmptyState, EmptyState, EvalClear, DoNotRecordCostModel)
	if err != nil {
		t.Fatal(err)
	}
	for i, symbol := range []*SymbolNode{a, b, c, d, e, captured} {
		want := []int32{1, 2, 7, 8, 9, 42}[i]
		if value, ok := ctx.Lookup(symbol).(*IntValue); !ok || value.Val != want {
			t.Fatalf("binding %s = %v, want %d", symbol, ctx.Lookup(symbol), want)
		}
	}
}

func TestToolFunctionContextTupleMismatchIsAnError(t *testing.T) {
	a, b, c := NewSymbolNode("a"), NewSymbolNode("b"), NewSymbolNode("c")
	short := NewTupleValue([]Value{NewIntValue(7)})
	params := NewFcnParams([][]*SymbolNode{{a}, {b, c}}, []bool{false, true}, []Value{
		NewIntervalValue(1, 1), NewSetEnumValue([]Value{short}, true),
	})
	tool := NewTool()
	fcn := NewFcnLambdaValue(params, BoolTrue, tool, EmptyContext, EmptyState, EmptyState, EvalClear)
	function := NewValueNode(fcn)
	function.Image = "F"
	expr := NewBuiltinOpApplNode(OpFA, function, NewTupleValue([]Value{NewIntValue(1), short}))
	_, err := tool.getFcnContext(fcn, expr, EmptyContext, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel)
	if err == nil || !strings.Contains(err.Error(), "the argument number 2 is:\n<<7>>which does not match its formal parameter.\nF") {
		t.Fatalf("tuple-mismatch error = %v", err)
	}
}

func TestToolFunctionContextSingleTupleMismatchPrintsToolIdentity(t *testing.T) {
	a, b := NewSymbolNode("a"), NewSymbolNode("b")
	short := NewTupleValue([]Value{NewIntValue(7)})
	tool := NewTool()
	fcn := NewFcnLambdaValue(NewTupleFcnParam([]*SymbolNode{a, b}, NewSetEnumValue([]Value{short}, true)), BoolTrue, tool, EmptyContext, EmptyState, EmptyState, EvalClear)
	function := NewValueNode(fcn)
	function.Image = "F"
	expr := NewBuiltinOpApplNode(OpFA, function, short)
	_, err := tool.getFcnContext(fcn, expr, EmptyContext, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel)
	// Preserve Java's this.toString() quirk with Go's runtime object identity.
	want := fmt.Sprintf("In applying the function\n%T@%p,\nthe argument is:\n<<7>>which does not match its formal parameter.\nF", tool, tool)
	if err == nil || err.Error() != want {
		t.Fatalf("single tuple mismatch = %v, want %q", err, want)
	}
}

func TestToolFunctionContextShortArgumentsKeepDirectIndexFailure(t *testing.T) {
	a, b := NewSymbolNode("a"), NewSymbolNode("b")
	tool := NewTool()
	fcn := NewFcnLambdaValue(NewFcnParams([][]*SymbolNode{{a, b}}, []bool{false}, []Value{NewIntervalValue(1, 2)}), BoolTrue, tool, EmptyContext, EmptyState, EmptyState, EvalClear)
	expr := NewBuiltinOpApplNode(OpFA, fcn, NewTupleValue([]Value{NewIntValue(1)}))
	defer func() {
		if err, ok := recover().(runtime.Error); !ok || !strings.Contains(err.Error(), "index out of range") {
			t.Fatalf("short function argument failure = %v, want direct-index failure", err)
		}
	}()
	_, _ = tool.getFcnContext(fcn, expr, EmptyContext, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel)
}
