package tlc

import "testing"

func TestFcnLambdaStringFallsBackAfterExpansionPanic(t *testing.T) {
	Globals.Lock()
	oldExpand := Globals.Expand
	Globals.Expand = true
	Globals.Unlock()
	t.Cleanup(func() {
		Globals.Lock()
		Globals.Expand = oldExpand
		Globals.Unlock()
	})
	x, y := NewSymbolNode("x"), NewSymbolNode("y")
	params := NewTupleFcnParam([]*SymbolNode{x, y}, NewSetEnumValue([]Value{NewTupleValue([]Value{NewIntValue(7)})}, true))
	fcn := NewFcnLambdaValue(params, BoolTrue, NewTool(), EmptyContext, EmptyState, EmptyState, EvalClear)
	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("lazy function formatting propagated expansion failure: %v", err)
		}
	}()
	want := "[<<x, y>> \\in {<<7>>} |-> <expression TRUE>]"
	if got := fcn.String(); got != want {
		t.Fatalf("lazy function string = %q, want %q", got, want)
	}
}
