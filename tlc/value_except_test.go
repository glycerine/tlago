package tlc

import "testing"

func TestValueExceptPublicCarrierSurface(t *testing.T) {
	path := []Value{IntOne, NewStringValue("x")}
	ex := NewValueExcept(path, NewIntValue(9))
	if ex.Index != 0 || ex.Current() != IntOne || ex.IsLast() {
		t.Fatalf("initial cursor = %d/%v/%t, want 0/1/false", ex.Index, ex.Current(), ex.IsLast())
	}
	copyAtTwo := NewValueExceptCopy(ex, 2)
	if &copyAtTwo.Path[0] != &ex.Path[0] || copyAtTwo.Value != ex.Value || copyAtTwo.Index != 2 {
		t.Fatal("indexed copy did not share path/value with the requested cursor")
	}
	copyAtTwo.Path[0] = IntZero
	if ex.Path[0] != IntZero {
		t.Fatal("copy did not retain the source path")
	}
	if got, want := copyAtTwo.String(), `.(2)0.(2)"x" = 9`; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	advanced := ex.Advanced()
	if advanced.Index != 1 || !advanced.IsLast() || ex.Index != 0 {
		t.Fatalf("advanced/original cursors = %d/%d, want 1/0", advanced.Index, ex.Index)
	}
}

func TestValueExceptCheckArgUsesFunctionDomains(t *testing.T) {
	single := &FcnLambdaValue{Params: NewSingleFcnParam(NewSymbolNode("x"), NewIntervalValue(1, 2))}
	ex := NewValueExcept([]Value{IntOne}, IntZero)
	if got, err := ex.CheckArg(single); err != nil || got != ex {
		t.Fatalf("single CheckArg = %p, %v; want receiver, nil", got, err)
	}
	outside := NewValueExcept([]Value{NewIntValue(3)}, IntZero)
	if got, err := outside.CheckArg(single); err != nil || got != nil {
		t.Fatalf("outside CheckArg = %p, %v; want nil, nil", got, err)
	}

	multi := &FcnLambdaValue{Params: NewFcnParams(
		[][]*SymbolNode{{NewSymbolNode("x")}, {NewSymbolNode("y"), NewSymbolNode("z")}},
		[]bool{false, false},
		[]Value{NewIntervalValue(1, 1), NewIntervalValue(2, 3)},
	)}
	tupleArg := NewTupleValue([]Value{IntOne, NewIntValue(2), NewIntValue(3)})
	tupleExcept := NewValueExcept([]Value{tupleArg}, IntZero)
	if got, err := tupleExcept.CheckArg(multi); err != nil || got != tupleExcept {
		t.Fatalf("multi CheckArg = %p, %v; want receiver, nil", got, err)
	}
}

func TestFcnParamsPublicCarrierAccessorsPreserveIdentity(t *testing.T) {
	formals := [][]*SymbolNode{{NewSymbolNode("x")}}
	domains := []Value{NewIntervalValue(1, 2)}
	params := NewFcnParams(formals, []bool{false}, domains)
	gotFormals, gotDomains := params.GetFormals(), params.GetDomains()
	if &gotFormals[0] != &formals[0] || &gotDomains[0] != &domains[0] {
		t.Fatal("FcnParams accessors did not return their backing slices")
	}
}
