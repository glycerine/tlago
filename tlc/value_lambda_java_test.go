/*******************************************************************************
 * Copyright (c) 2026 NVIDIA Corp. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Contributors:
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/

// Complete original FcnLambdaValueTest, retaining each original assertion/call.
// Java null CostModel arguments use the native zero carrier (Node == nil).
package tlc

import "testing"

func javaCreateFcnLambda(domain, val Value) *FcnLambdaValue {
	params := NewFcnParams([][]*SymbolNode{{NewFormalParamSymbolNode("x", 0)}}, []bool{false}, []Value{domain})
	// TLCStates.createDummyState(): a single variable v0, bound to 0, uid 0.
	SetStateVariables([]string{"v0"})
	state := NewEmptyState()
	state.UID = 0
	state.Bind(UniqueStringOf("v0"), NewIntValue(0))
	tool := &Tool{EvalFunc: func(*Tool, SemanticNode, ...any) (Value, error) { return val, nil }}
	return NewFcnLambdaValue(params, NullSemanticNodeInstance, tool, EmptyContext, state, nil, EvalClear)
}
func javaFcnLambdaSelect(t *testing.T, f *FcnLambdaValue, arg Value) Value {
	t.Helper()
	v, err := f.Select(arg)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func javaFcnLambdaApply(t *testing.T, f *FcnLambdaValue, arg Value, control int) Value {
	t.Helper()
	v, err := f.ApplyWithControl(arg, control)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func javaFcnLambdaExcept(t *testing.T, f *FcnLambdaValue, ex *ValueExcept) Value {
	t.Helper()
	v, err := f.TakeExcept(ex)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func javaFcnLambdaDomain(t *testing.T, f *FcnLambdaValue) Value {
	t.Helper()
	v, err := f.GetDomain()
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func javaFcnLambdaEquals(t *testing.T, expected, actual any) {
	t.Helper()
	if expectedValue, ok := expected.(Value); ok {
		actualValue, ok := actual.(Value)
		if !ok {
			t.Fatalf("actual %T, want Value", actual)
		}
		eq, err := expectedValue.Equal(actualValue)
		if err != nil || !eq {
			t.Fatalf("equals=%v/%v, want true", eq, err)
		}
	} else if expected != actual {
		t.Fatalf("got %v, want %v", actual, expected)
	}
}
func TestJavaFcnLambdaValue(t *testing.T) {
	oldIntern, oldVars, oldEmpty := internTable, stateVariables, EmptyState
	oldTool, oldMetadata, oldSymmetry := stateTool, statePreserveMetadata, stateSymmetryPermutations
	oldPoly := FP64IrredPoly()
	internTable = NewInternTable(1024)
	stateTool = nil
	statePreserveMetadata = false
	stateSymmetryPermutations = nil
	t.Cleanup(func() {
		internTable, stateVariables, EmptyState = oldIntern, oldVars, oldEmpty
		stateTool, statePreserveMetadata, stateSymmetryPermutations = oldTool, oldMetadata, oldSymmetry
		FP64InitPoly(oldPoly)
	})

	t.Run("testEmptyIntervalDomainToTuple", func(t *testing.T) {
		fcn := javaCreateFcnLambda(NewIntervalValue(2, 1), NewIntValue(42))
		javaFcnLambdaEquals(t, EmptyTuple, fcn.ToTuple())
		if v := javaCreateFcnLambda(NewIntervalValue(2, 2), NewIntValue(42)).ToTuple(); v != nil {
			t.Fatalf("expected null, got %v", v)
		}
	})
	t.Run("testToString", func(t *testing.T) {
		flv := javaCreateFcnLambda(NewIntervalValue(1, 2), NewIntValue(42))
		javaFcnLambdaEquals(t, "<<42, 42>>", flv.String())
		flv = javaCreateFcnLambda(NewSetEnumValue([]Value{NewStringValue("a"), NewStringValue("b")}, false), NewIntValue(42))
		javaFcnLambdaEquals(t, "[a |-> 42, b |-> 42]", flv.String())
	})
	t.Run("testToFcnRcd", func(t *testing.T) {
		flv := javaCreateFcnLambda(NewIntervalValue(1, 2), NewIntValue(42))
		javaFcnLambdaEquals(t, "<<42, 42>>", flv.ToFcnRcd().String())
		flv = javaCreateFcnLambda(NewSetEnumValue([]Value{NewStringValue("a"), NewStringValue("b")}, false), NewIntValue(42))
		javaFcnLambdaEquals(t, "[a |-> 42, b |-> 42]", flv.ToFcnRcd().String())
	})
	t.Run("testSelectAndApply", func(t *testing.T) {
		flv := javaCreateFcnLambda(NewIntervalValue(1, 3), NewIntValue(7))
		javaFcnLambdaEquals(t, NewIntValue(7), javaFcnLambdaSelect(t, flv, NewIntValue(1)))
		javaFcnLambdaEquals(t, NewIntValue(7), javaFcnLambdaSelect(t, flv, NewIntValue(2)))
		javaFcnLambdaEquals(t, NewIntValue(7), javaFcnLambdaApply(t, flv, NewIntValue(3), EvalClear))
		javaFcnLambdaEquals(t, "7", javaFcnLambdaSelect(t, flv, NewIntValue(1)).String())
		javaFcnLambdaEquals(t, "7", javaFcnLambdaApply(t, flv, NewIntValue(3), EvalClear).String())
		javaFcnLambdaEquals(t, "<<7, 7, 7>>", flv.String())
	})
	t.Run("testTakeExceptOverridesValue", func(t *testing.T) {
		base := javaCreateFcnLambda(NewIntervalValue(1, 3), NewIntValue(7))
		ex := &ValueExcept{Path: []Value{NewIntValue(2)}, Value: NewIntValue(99)}
		updated := javaFcnLambdaExcept(t, base, ex).(*FcnLambdaValue)
		javaFcnLambdaEquals(t, NewIntValue(7), javaFcnLambdaSelect(t, updated, NewIntValue(1)))
		javaFcnLambdaEquals(t, NewIntValue(99), javaFcnLambdaSelect(t, updated, NewIntValue(2)))
		javaFcnLambdaEquals(t, NewIntValue(7), javaFcnLambdaSelect(t, updated, NewIntValue(3)))
		javaFcnLambdaEquals(t, "<<7, 7, 7>>", base.String())
		javaFcnLambdaEquals(t, "<<7, 99, 7>>", updated.String())
	})
	t.Run("testDomainIsStableAcrossConversion", func(t *testing.T) {
		flv := javaCreateFcnLambda(NewIntervalValue(1, 2), NewIntValue(42))
		javaFcnLambdaEquals(t, NewIntervalValue(1, 2), javaFcnLambdaDomain(t, flv))
		javaFcnLambdaEquals(t, "1..2", javaFcnLambdaDomain(t, flv).String())
		javaFcnLambdaEquals(t, "<<42, 42>>", flv.ToFcnRcd().String())
		javaFcnLambdaEquals(t, NewIntervalValue(1, 2), javaFcnLambdaDomain(t, flv))
		javaFcnLambdaEquals(t, "1..2", javaFcnLambdaDomain(t, flv).String())
	})
	t.Run("testToFcnRcdReturnsCachedInstance", func(t *testing.T) {
		flv := javaCreateFcnLambda(NewIntervalValue(1, 2), NewIntValue(42))
		first := flv.ToFcnRcd()
		second := flv.ToFcnRcd()
		if first != second {
			t.Fatal("expected identical instances")
		}
		javaFcnLambdaEquals(t, "<<42, 42>>", first.String())
		javaFcnLambdaEquals(t, "<<42, 42>>", second.String())
	})
	t.Run("testToRcdForIntervalDomainIsNull", func(t *testing.T) {
		flv := javaCreateFcnLambda(NewIntervalValue(1, 2), NewIntValue(42))
		if v := flv.ToRecord(); v != nil {
			t.Fatalf("expected null, got %v", v)
		}
	})
	t.Run("testToRcdForStringDomain", func(t *testing.T) {
		flv := javaCreateFcnLambda(NewSetEnumValue([]Value{NewStringValue("a"), NewStringValue("b")}, false), NewIntValue(5))
		javaFcnLambdaEquals(t, "[a |-> 5, b |-> 5]", flv.String())
		javaFcnLambdaEquals(t, "[a |-> 5, b |-> 5]", flv.ToRecord().String())
	})
	t.Run("testFingerprintStableAcrossConversion", func(t *testing.T) {
		intervalFcn := javaCreateFcnLambda(NewIntervalValue(1, 3), NewIntValue(9))
		javaFcnLambdaEquals(t, "<<9, 9, 9>>", intervalFcn.String())
		fpIntervalBefore := intervalFcn.FingerPrint(0)
		javaFcnLambdaEquals(t, "<<9, 9, 9>>", intervalFcn.ToFcnRcd().String())
		javaFcnLambdaEquals(t, "<<9, 9, 9>>", intervalFcn.String())
		if v := intervalFcn.ToRecord(); v != nil {
			t.Fatalf("expected null, got %v", v)
		}
		fpIntervalAfter := intervalFcn.FingerPrint(0)
		javaFcnLambdaEquals(t, fpIntervalBefore, fpIntervalAfter)
		stringFcn := javaCreateFcnLambda(NewSetEnumValue([]Value{NewStringValue("a"), NewStringValue("b")}, false), NewIntValue(9))
		javaFcnLambdaEquals(t, "[a |-> 9, b |-> 9]", stringFcn.String())
		fpStringBefore := stringFcn.FingerPrint(0)
		javaFcnLambdaEquals(t, "[a |-> 9, b |-> 9]", stringFcn.ToFcnRcd().String())
		javaFcnLambdaEquals(t, "[a |-> 9, b |-> 9]", stringFcn.ToRecord().String())
		javaFcnLambdaEquals(t, "[a |-> 9, b |-> 9]", stringFcn.String())
		fpStringAfter := stringFcn.FingerPrint(0)
		javaFcnLambdaEquals(t, fpStringBefore, fpStringAfter)
	})
	t.Run("testTakeExceptThenConvertToFcnRcd", func(t *testing.T) {
		base := javaCreateFcnLambda(NewIntervalValue(1, 2), NewIntValue(1))
		updated := javaFcnLambdaExcept(t, base, &ValueExcept{Path: []Value{NewIntValue(1)}, Value: NewIntValue(8)}).(*FcnLambdaValue)
		javaFcnLambdaEquals(t, "<<8, 1>>", updated.ToFcnRcd().String())
	})
	t.Run("testToFcnRcdAssertFail", func(t *testing.T) {
		base := javaCreateFcnLambda(NewIntervalValue(1, 2), NewIntValue(42))
		ex1 := &ValueExcept{Path: []Value{NewIntValue(1)}, Value: NewIntValue(99)}
		flv1 := javaFcnLambdaExcept(t, base, ex1).(*FcnLambdaValue)
		ex2 := &ValueExcept{Path: []Value{NewIntValue(2)}, Value: NewIntValue(100)}
		flv2 := javaFcnLambdaExcept(t, flv1, ex2).(*FcnLambdaValue)
		flv1.ToFcnRcd()
		flv2.ToFcnRcd()
		javaFcnLambdaEquals(t, NewIntValue(99), javaFcnLambdaSelect(t, flv1, NewIntValue(1)))
		javaFcnLambdaEquals(t, NewIntValue(42), javaFcnLambdaSelect(t, flv1, NewIntValue(2)))
		javaFcnLambdaEquals(t, NewIntValue(99), javaFcnLambdaSelect(t, flv2, NewIntValue(1)))
		javaFcnLambdaEquals(t, NewIntValue(100), javaFcnLambdaSelect(t, flv2, NewIntValue(2)))
	})
	t.Run("testToFcnRcdClassCastException", func(t *testing.T) {
		base := javaCreateFcnLambda(NewIntervalValue(1, 2), NewIntValue(42))
		ex1 := &ValueExcept{Path: []Value{NewIntValue(1)}, Value: NewIntValue(99)}
		flv1 := javaFcnLambdaExcept(t, base, ex1).(*FcnLambdaValue)
		ex2 := &ValueExcept{Path: []Value{NewIntValue(2)}, Value: NewIntValue(100)}
		flv2 := javaFcnLambdaExcept(t, flv1, ex2).(*FcnLambdaValue)
		flv2.ToFcnRcd()
		flv1.ToFcnRcd()
		javaFcnLambdaEquals(t, NewIntValue(99), javaFcnLambdaSelect(t, flv1, NewIntValue(1)))
		javaFcnLambdaEquals(t, NewIntValue(42), javaFcnLambdaSelect(t, flv1, NewIntValue(2)))
		javaFcnLambdaEquals(t, NewIntValue(99), javaFcnLambdaSelect(t, flv2, NewIntValue(1)))
		javaFcnLambdaEquals(t, NewIntValue(100), javaFcnLambdaSelect(t, flv2, NewIntValue(2)))
	})
	t.Run("testToFcnRcdSilentCorruption", func(t *testing.T) {
		defaultFcn := NewFcnRcdIntervalValue(NewIntervalValue(1, 2), []Value{NewIntValue(0), NewIntValue(0)}, CostModel{})
		base := javaCreateFcnLambda(NewIntervalValue(1, 2), defaultFcn)
		val1 := NewFcnRcdIntervalValue(NewIntervalValue(1, 2), []Value{NewIntValue(11), NewIntValue(12)}, CostModel{})
		ex1 := &ValueExcept{Path: []Value{NewIntValue(1)}, Value: val1}
		flv1 := javaFcnLambdaExcept(t, base, ex1).(*FcnLambdaValue)
		val2 := NewFcnRcdIntervalValue(NewIntervalValue(1, 2), []Value{NewIntValue(21), NewIntValue(22)}, CostModel{})
		ex2 := &ValueExcept{Path: []Value{NewIntValue(2)}, Value: val2}
		flv2 := javaFcnLambdaExcept(t, flv1, ex2).(*FcnLambdaValue)
		javaFcnLambdaEquals(t, "<<<<11, 12>>, <<0, 0>>>>", flv1.ToFcnRcd().String())
		javaFcnLambdaEquals(t, "<<<<11, 12>>, <<21, 22>>>>", flv2.ToFcnRcd().String())
	})
	t.Run("testToFcnRcdSilentCorruptionFP", func(t *testing.T) {
		defaultFcn := NewFcnRcdIntervalValue(NewIntervalValue(1, 2), []Value{NewIntValue(0), NewIntValue(0)}, CostModel{})
		base := javaCreateFcnLambda(NewIntervalValue(1, 2), defaultFcn)
		val1 := NewFcnRcdIntervalValue(NewIntervalValue(1, 2), []Value{NewIntValue(11), NewIntValue(12)}, CostModel{})
		ex1 := &ValueExcept{Path: []Value{NewIntValue(1)}, Value: val1}
		flv1 := javaFcnLambdaExcept(t, base, ex1).(*FcnLambdaValue)
		val2 := NewFcnRcdIntervalValue(NewIntervalValue(1, 2), []Value{NewIntValue(21), NewIntValue(22)}, CostModel{})
		ex2 := &ValueExcept{Path: []Value{NewIntValue(2)}, Value: val2}
		flv2 := javaFcnLambdaExcept(t, flv1, ex2).(*FcnLambdaValue)
		flv1.ToFcnRcd()
		fcnRcd := flv2.ToFcnRcd()
		freshY := NewFcnRcdIntervalValue(NewIntervalValue(1, 2), []Value{val1, val2}, CostModel{})
		FP64Init()
		javaFcnLambdaEquals(t, freshY.FingerPrint(0), fcnRcd.FingerPrint(0))
	})
	t.Run("testToTupleWithExceptIntervalDomain", func(t *testing.T) {
		base := javaCreateFcnLambda(NewIntervalValue(1, 3), NewIntValue(42))
		ex := &ValueExcept{Path: []Value{NewIntValue(2)}, Value: NewIntValue(99)}
		updated := javaFcnLambdaExcept(t, base, ex).(*FcnLambdaValue)
		javaFcnLambdaEquals(t, "<<42, 99, 42>>", updated.ToTuple().String())
	})
	t.Run("testToTupleWithExceptSetEnumDomain", func(t *testing.T) {
		domain := NewSetEnumValue([]Value{NewIntValue(1), NewIntValue(2), NewIntValue(3)}, true)
		base := javaCreateFcnLambda(domain, NewIntValue(42))
		ex := &ValueExcept{Path: []Value{NewIntValue(2)}, Value: NewIntValue(99)}
		updated := javaFcnLambdaExcept(t, base, ex).(*FcnLambdaValue)
		javaFcnLambdaEquals(t, "<<42, 99, 42>>", updated.ToTuple().String())
	})
	t.Run("testToTupleWithExceptFP", func(t *testing.T) {
		base := javaCreateFcnLambda(NewIntervalValue(1, 3), NewIntValue(42))
		ex := &ValueExcept{Path: []Value{NewIntValue(2)}, Value: NewIntValue(99)}
		updated := javaFcnLambdaExcept(t, base, ex).(*FcnLambdaValue)
		tuple := updated.ToTuple()
		correct := NewTupleValue([]Value{NewIntValue(42), NewIntValue(99), NewIntValue(42)})
		FP64Init()
		javaFcnLambdaEquals(t, correct.FingerPrint(0), tuple.FingerPrint(0))
	})
}
