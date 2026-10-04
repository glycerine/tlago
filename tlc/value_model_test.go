/*******************************************************************************
 * Copyright (c) 2021 Microsoft Research. All rights reserved.
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

// Complete ModelValueTest translation in JUnit 4's original runner order.
// Shared intern IDs affect exact comparison results, so preserve the order and
// evaluate every constructor at its original assertion/call site.
package tlc

import "testing"

func TestJavaModelValue(t *testing.T) {
	oldIntern := internTable
	modelValues.Lock()
	oldCount, oldTable, oldMVs := modelValues.count, modelValues.table, modelValues.mvs
	modelValues.Unlock()
	internTable = NewInternTable(1024)
	ModelValueInit()
	t.Cleanup(func() {
		internTable = oldIntern
		modelValues.Lock()
		modelValues.count, modelValues.table, modelValues.mvs = oldCount, oldTable, oldMVs
		modelValues.Unlock()
	})

	t.Run("testEqualsRcdTypedMV", func(t *testing.T) {
		if _, err := NewRecordValue([]*UniqueString{UniqueStringOf("foo")}, []Value{NewStringValue("bar")}, false).Equal(MakeModelValue("B_b")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testCompareToRcdTypedMV", func(t *testing.T) {
		if _, err := NewRecordValue([]*UniqueString{UniqueStringOf("foo")}, []Value{NewStringValue("bar")}, false).Compare(MakeModelValue("B_b")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testCompareToIntMV", func(t *testing.T) {
		if got, err := NewIntValue(0).Compare(MakeModelValue("untyped")); err != nil || got != 1 {
			t.Fatalf("compare=%d/%v, want 1", got, err)
		}
		if got, err := MakeModelValue("untyped").Compare(NewIntValue(0)); err != nil || got != -1 {
			t.Fatalf("compare=%d/%v, want -1", got, err)
		}
	})
	t.Run("testCompareToRcdMV", func(t *testing.T) {
		if got, err := NewRecordValue([]*UniqueString{UniqueStringOf("foo")}, []Value{NewStringValue("bar")}, false).Compare(MakeModelValue("untyped")); err != nil || got != 1 {
			t.Fatalf("compare=%d/%v, want 1", got, err)
		}
		if got, err := MakeModelValue("untyped").Compare(NewRecordValue([]*UniqueString{UniqueStringOf("foo")}, []Value{NewStringValue("bar")}, false)); err != nil || got != -1 {
			t.Fatalf("compare=%d/%v, want -1", got, err)
		}
	})
	t.Run("testCompareToTyped", func(t *testing.T) {
		a := MakeModelValue("A_a")
		b := MakeModelValue("A_b")
		if got, err := a.Compare(b); err != nil || got != -1 {
			t.Fatalf("compare=%d/%v, want -1", got, err)
		}
		if got, err := b.Compare(a); err != nil || got != 1 {
			t.Fatalf("compare=%d/%v, want 1", got, err)
		}
		if got, err := a.Compare(a); err != nil || got != 0 {
			t.Fatalf("compare=%d/%v, want 0", got, err)
		}
	})
	t.Run("testCompareToTypedMVsUntyped", func(t *testing.T) {
		if got, err := MakeModelValue("A_a").Compare(MakeModelValue("untyped")); err != nil || got != 1 {
			t.Fatalf("compare=%d/%v, want 1", got, err)
		}
		if got, err := MakeModelValue("untyped").Compare(MakeModelValue("A_a")); err != nil || got != -1 {
			t.Fatalf("compare=%d/%v, want -1", got, err)
		}
	})
	t.Run("testEqualsIntMV", func(t *testing.T) {
		if got, err := NewIntValue(0).Equal(MakeModelValue("untyped")); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
		if got, err := MakeModelValue("untyped").Equal(NewIntValue(0)); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
	})
	t.Run("testEqualsRcdMV", func(t *testing.T) {
		if got, err := NewRecordValue([]*UniqueString{UniqueStringOf("foo")}, []Value{NewStringValue("bar")}, false).Equal(MakeModelValue("untyped")); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
		if got, err := MakeModelValue("untyped").Equal(NewRecordValue([]*UniqueString{UniqueStringOf("foo")}, []Value{NewStringValue("bar")}, false)); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
	})
	t.Run("testEqualsTyped", func(t *testing.T) {
		a := MakeModelValue("A_a")
		b := MakeModelValue("A_b")
		if got, err := a.Equal(b); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
		if got, err := b.Equal(a); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
		if got, err := a.Equal(a); err != nil || got != true {
			t.Fatalf("equals=%v/%v, want true", got, err)
		}
	})
	t.Run("testCompareToBoolMV", func(t *testing.T) {
		if got, err := BoolFalse.Compare(MakeModelValue("untyped")); err != nil || got != 1 {
			t.Fatalf("compare=%d/%v, want 1", got, err)
		}
		if got, err := MakeModelValue("untyped").Compare(BoolFalse); err != nil || got != -1 {
			t.Fatalf("compare=%d/%v, want -1", got, err)
		}
	})
	t.Run("testEqualsIntTypedMV", func(t *testing.T) {
		if _, err := NewIntValue(0).Equal(MakeModelValue("B_b")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testCompareToTypedMVString", func(t *testing.T) {
		if _, err := MakeModelValue("B_b").Compare(NewStringValue("str")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testCompareToIntTypedMV", func(t *testing.T) {
		if _, err := NewIntValue(0).Compare(MakeModelValue("B_b")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testCompareToTupeMV", func(t *testing.T) {
		if got, err := NewStringValue("foo").Compare(MakeModelValue("untyped")); err != nil || got != 1 {
			t.Fatalf("compare=%d/%v, want 1", got, err)
		}
		if got, err := MakeModelValue("untyped").Compare(NewStringValue("foo")); err != nil || got != -1 {
			t.Fatalf("compare=%d/%v, want -1", got, err)
		}
	})
	t.Run("testEqualsBoolMV", func(t *testing.T) {
		if got, err := BoolFalse.Equal(MakeModelValue("untyped")); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
		if got, err := MakeModelValue("untyped").Equal(BoolFalse); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
	})
	t.Run("testCompareToStringTypedMV", func(t *testing.T) {
		if _, err := NewStringValue("str").Compare(MakeModelValue("B_b")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testCompareToIVTypedMV", func(t *testing.T) {
		if _, err := NewIntervalValue(0, 1).Compare(MakeModelValue("B_b")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testEqualsIVMV", func(t *testing.T) {
		if got, err := NewIntervalValue(0, 1).Equal(MakeModelValue("untyped")); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
		if got, err := MakeModelValue("untyped").Equal(NewIntervalValue(0, 1)); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
	})
	t.Run("testCompareToTypedMVBool", func(t *testing.T) {
		if _, err := MakeModelValue("B_b").Compare(BoolTrue); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testCompareToTypedMVTupe", func(t *testing.T) {
		if _, err := MakeModelValue("B_b").Compare(NewTupleValue([]Value{NewStringValue("foo")})); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testEqualsIVTypedMV", func(t *testing.T) {
		if _, err := NewIntervalValue(0, 1).Equal(MakeModelValue("B_b")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testEqualsTypedMVBool", func(t *testing.T) {
		if _, err := MakeModelValue("B_b").Equal(BoolTrue); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testEqualsTypedMVTupe", func(t *testing.T) {
		if _, err := MakeModelValue("B_b").Equal(NewTupleValue([]Value{NewStringValue("foo")})); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testCompareToStringMV", func(t *testing.T) {
		if got, err := NewStringValue("str").Compare(MakeModelValue("untyped")); err != nil || got != 1 {
			t.Fatalf("compare=%d/%v, want 1", got, err)
		}
		if got, err := MakeModelValue("untyped").Compare(NewStringValue("str")); err != nil || got != -1 {
			t.Fatalf("compare=%d/%v, want -1", got, err)
		}
	})
	t.Run("testEqualsTupeMV", func(t *testing.T) {
		if got, err := NewTupleValue([]Value{NewStringValue("foo")}).Equal(MakeModelValue("untyped")); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
		if got, err := MakeModelValue("untyped").Equal(NewTupleValue([]Value{NewStringValue("foo")})); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
	})
	t.Run("testCompareToUntyped", func(t *testing.T) {
		u1 := MakeModelValue("u1")
		u2 := MakeModelValue("u2")
		if got, err := u1.Compare(u2); err != nil || got != -1 {
			t.Fatalf("compare=%d/%v, want -1", got, err)
		}
		if got, err := u2.Compare(u1); err != nil || got != 1 {
			t.Fatalf("compare=%d/%v, want 1", got, err)
		}
		if got, err := u1.Compare(u1); err != nil || got != 0 {
			t.Fatalf("compare=%d/%v, want 0", got, err)
		}
	})
	t.Run("testEqualsUntyped", func(t *testing.T) {
		u1 := MakeModelValue("u1")
		u2 := MakeModelValue("u2")
		if got, err := u1.Equal(u2); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
		if got, err := u2.Equal(u1); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
		if got, err := u1.Equal(u1); err != nil || got != true {
			t.Fatalf("equals=%v/%v, want true", got, err)
		}
	})
	t.Run("testCompareToTupeTypedMV", func(t *testing.T) {
		if _, err := NewTupleValue([]Value{NewStringValue("foo")}).Compare(MakeModelValue("B_b")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testEqualsStringMV", func(t *testing.T) {
		if got, err := NewStringValue("str").Equal(MakeModelValue("untyped")); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
		if got, err := MakeModelValue("untyped").Equal(NewStringValue("str")); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
	})
	t.Run("testEqualsTupeTypedMV", func(t *testing.T) {
		if _, err := NewTupleValue([]Value{NewStringValue("foo")}).Equal(MakeModelValue("B_b")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testCompareToTwoTypedMVs", func(t *testing.T) {
		if _, err := MakeModelValue("A_a").Compare(MakeModelValue("B_b")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testCompareToTypedMVIV", func(t *testing.T) {
		if _, err := MakeModelValue("B_b").Compare(NewIntervalValue(0, 1)); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testEqualsTypedMVsUntyped", func(t *testing.T) {
		if got, err := MakeModelValue("A_a").Equal(MakeModelValue("untyped")); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
		if got, err := MakeModelValue("untyped").Equal(MakeModelValue("A_a")); err != nil || got != false {
			t.Fatalf("equals=%v/%v, want false", got, err)
		}
	})
	t.Run("testEqualsTypedMVString", func(t *testing.T) {
		if _, err := MakeModelValue("B_b").Equal(NewStringValue("str")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testEqualsTwoTypedMVs", func(t *testing.T) {
		if _, err := MakeModelValue("A_a").Equal(MakeModelValue("B_b")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testEqualsTypedMVIV", func(t *testing.T) {
		if _, err := MakeModelValue("B_b").Equal(NewIntervalValue(0, 1)); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testCompareToIVMV", func(t *testing.T) {
		if got, err := NewIntervalValue(0, 1).Compare(MakeModelValue("untyped")); err != nil || got != 1 {
			t.Fatalf("compare=%d/%v, want 1", got, err)
		}
		if got, err := MakeModelValue("untyped").Compare(NewIntervalValue(0, 1)); err != nil || got != -1 {
			t.Fatalf("compare=%d/%v, want -1", got, err)
		}
	})
	t.Run("testCompareToBoolTypedMV", func(t *testing.T) {
		if _, err := BoolFalse.Compare(MakeModelValue("B_b")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testEqualsBoolTypedMV", func(t *testing.T) {
		if _, err := BoolFalse.Equal(MakeModelValue("B_b")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testEqualsStringTypedMV", func(t *testing.T) {
		if _, err := NewStringValue("str").Equal(MakeModelValue("B_b")); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testEqualsTypedMVInt", func(t *testing.T) {
		if _, err := MakeModelValue("B_b").Equal(NewIntValue(0)); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testEqualsTypedMVRcd", func(t *testing.T) {
		if _, err := MakeModelValue("B_b").Equal(NewRecordValue([]*UniqueString{UniqueStringOf("foo")}, []Value{NewStringValue("bar")}, false)); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testCompareToTypedMVInt", func(t *testing.T) {
		if _, err := MakeModelValue("B_b").Compare(NewIntValue(0)); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testCompareToTypedMVRcd", func(t *testing.T) {
		if _, err := MakeModelValue("B_b").Compare(NewRecordValue([]*UniqueString{UniqueStringOf("foo")}, []Value{NewStringValue("bar")}, false)); javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
}
