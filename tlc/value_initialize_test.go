/*******************************************************************************
 * Copyright (c) 2019 Microsoft Research. All rights reserved.
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
// Complete translation of all ten original InitializeValueTest methods.
// Preserve shared ValueVec instances, original inputs, and assertion order.
package tlc

import (
	"fmt"
	"testing"
)

func javaInitializeValueSetup(t *testing.T) {
	t.Helper()
	previous := FP64IrredPoly()
	FP64Init()
	t.Cleanup(func() { FP64InitPoly(previous) })
}

func requireJavaInitialization(t *testing.T, value Value, expected bool) {
	t.Helper()
	if got := value.IsNormalized(); got != expected {
		t.Fatalf("%T.isNormalized = %v, want %v", value, got, expected)
	}
}

func javaInitializeValueVec() *ValueVec {
	vec := NewValueVec(10)
	vec.Add(NewIntValue(42))
	vec.Add(NewIntValue(23))
	vec.Add(NewIntValue(4711))
	vec.Add(NewIntValue(1))
	return vec
}

func TestJavaInitializeValueUnion(t *testing.T) {
	javaInitializeValueSetup(t)
	vec := NewValueVec(10)
	vec.Add(NewSetEnumValue([]Value{NewIntValue(42)}, false))
	vec.Add(NewSetEnumValue([]Value{NewIntValue(23)}, false))
	vec.Add(NewSetEnumValue([]Value{NewIntValue(4711)}, false))
	vec.Add(NewSetEnumValue([]Value{NewIntValue(1)}, false))
	uv := NewUnionValue(NewSetEnumValueVec(vec, false))
	requireJavaInitialization(t, uv, false)
	requireJavaInitialization(t, uv.Set, false)
	InitializeValue(uv)
	requireJavaInitialization(t, uv.Set, true)
	requireJavaInitialization(t, uv, true)
}

func TestJavaInitializeValueSetCap(t *testing.T) {
	javaInitializeValueSetup(t)
	vec := javaInitializeValueVec()
	scv := NewSetCapValue(NewSetEnumValueVec(vec, false), NewSetEnumValueVec(vec, false))
	requireJavaInitialization(t, scv, false)
	requireJavaInitialization(t, scv.Set1, false)
	requireJavaInitialization(t, scv.Set2, false)
	InitializeValue(scv)
	requireJavaInitialization(t, scv.Set1, true)
	requireJavaInitialization(t, scv.Set2, true)
	requireJavaInitialization(t, scv, true)
}

func TestJavaInitializeValueSetCup(t *testing.T) {
	javaInitializeValueSetup(t)
	vec := javaInitializeValueVec()
	scv := NewSetCupValue(NewSetEnumValueVec(vec, false), NewSetEnumValueVec(vec, false))
	requireJavaInitialization(t, scv, false)
	requireJavaInitialization(t, scv.Set1, false)
	requireJavaInitialization(t, scv.Set2, false)
	InitializeValue(scv)
	requireJavaInitialization(t, scv.Set1, true)
	requireJavaInitialization(t, scv.Set2, true)
	requireJavaInitialization(t, scv, true)
}

func TestJavaInitializeValueSetDiff(t *testing.T) {
	javaInitializeValueSetup(t)
	vec := javaInitializeValueVec()
	scv := NewSetDiffValue(NewSetEnumValueVec(vec, false), NewSetEnumValueVec(vec, false))
	requireJavaInitialization(t, scv, false)
	requireJavaInitialization(t, scv.Set1, false)
	requireJavaInitialization(t, scv.Set2, false)
	InitializeValue(scv)
	requireJavaInitialization(t, scv.Set1, true)
	requireJavaInitialization(t, scv.Set2, true)
	requireJavaInitialization(t, scv, true)
}

func TestJavaInitializeValueSubset(t *testing.T) {
	javaInitializeValueSetup(t)
	vec := javaInitializeValueVec()
	scv := NewSubsetValue(NewSetEnumValueVec(vec, false))
	requireJavaInitialization(t, scv, false)
	requireJavaInitialization(t, scv.Set, false)
	InitializeValue(scv)
	requireJavaInitialization(t, scv.Set, true)
	requireJavaInitialization(t, scv, true)
}

func TestJavaInitializeValueRecord(t *testing.T) {
	javaInitializeValueSetup(t)
	internTable := NewInternTable(2)
	a, b := internTable.Put("a"), internTable.Put("b")
	vec := javaInitializeValueVec()
	aVal, bVal := NewSetEnumValueVec(vec, false), NewSetEnumValueVec(vec, false)
	rcdv := NewRecordValue([]*UniqueString{b, a}, []Value{bVal, aVal}, false)
	requireJavaInitialization(t, rcdv, false)
	for _, v := range rcdv.Values {
		requireJavaInitialization(t, v, false)
	}
	InitializeValue(rcdv)
	for _, v := range rcdv.Values {
		requireJavaInitialization(t, v, true)
	}
	requireJavaInitialization(t, rcdv, true)
}

func TestJavaInitializeValueFcnRecord(t *testing.T) {
	javaInitializeValueSetup(t)
	vec := javaInitializeValueVec()
	aVal, bVal := NewSetEnumValueVec(vec, false), NewSetEnumValueVec(vec, false)
	rcdv := NewFcnRcdValue([]Value{NewStringValue("B"), NewStringValue("A")}, []Value{bVal, aVal}, false)
	requireJavaInitialization(t, rcdv, false)
	for _, v := range rcdv.Values {
		requireJavaInitialization(t, v, false)
	}
	InitializeValue(rcdv)
	for _, v := range rcdv.Values {
		requireJavaInitialization(t, v, true)
	}
	requireJavaInitialization(t, rcdv, true)
}

func TestJavaInitializeValueTuple(t *testing.T) {
	javaInitializeValueSetup(t)
	vec := javaInitializeValueVec()
	aVal := NewSetEnumValueVec(vec, false)
	tuple := NewTupleValue([]Value{aVal})
	InitializeValue(tuple)
	requireJavaInitialization(t, tuple, true)
	for _, v := range tuple.Elems {
		requireJavaInitialization(t, v, true)
	}
}

func TestJavaInitializeValueSetOfTuple(t *testing.T) {
	javaInitializeValueSetup(t)
	intVal := NewIntervalValue(1, 2)
	inner := NewSetOfTuplesValue([]Value{intVal, intVal})
	tuples := NewSetOfTuplesValue([]Value{inner, inner})
	InitializeValue(tuples)
	requireJavaInitialization(t, tuples, true)
	for _, v := range tuples.Sets {
		requireJavaInitialization(t, v, true)
	}
}

func TestJavaInitializeValueSetOfRcds(t *testing.T) {
	javaInitializeValueSetup(t)
	values := []Value{
		NewSetEnumValue(javaInitializeRecordValues(7, "a"), true),
		NewIntervalValue(1, 2),
		NewIntervalValue(1, 4),
	}
	setOfRcrds, err := NewSetOfRcdsValue(javaInitializeRecordNames(3), values, true)
	if err != nil {
		t.Fatal(err)
	}
	InitializeValue(setOfRcrds)
	requireJavaInitialization(t, setOfRcrds, true)
	for _, v := range setOfRcrds.Values {
		requireJavaInitialization(t, v, true)
	}
}

func javaInitializeRecordValues(n int, str string) []Value {
	values := make([]Value, n)
	for i := range values {
		values[i] = NewStringValue(fmt.Sprint(str, i))
	}
	return values
}

func javaInitializeRecordNames(n int) []*UniqueString {
	names := make([]*UniqueString, n)
	for i := range names {
		names[i] = UniqueStringOf(fmt.Sprint("N", i))
	}
	return names
}
