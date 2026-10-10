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
// Complete original KSubsetValueTest, including all thirty fingerprint rows.
package tlc

import "testing"

func javaKSubsetDoTest(t *testing.T, ints []int, iv *IntervalValue) {
	t.Helper()
	for _, i := range ints {
		expected := Choose(javaSubsetEnumerationSize(t, iv), i)
		value := NewKSubsetValue(i, iv)
		javaFcnSetSize(t, value, int(expected))
		value.FingerPrint(0) // FP64.Zero, as in source.
		javaFcnSetSize(t, value, int(expected))
		set, err := value.ToSetEnum()
		if err != nil {
			t.Fatal(err)
		}
		javaFcnSetSize(t, set, int(expected))
	}
}
func TestJavaKSubsetValue(t *testing.T) {
	oldIntern := internTable
	internTable = NewInternTable(1024)
	t.Cleanup(func() { internTable = oldIntern })
	t.Run("testEnumerateN32", func(t *testing.T) {
		iv := NewIntervalValue(1, 32)
		javaFcnSetSize(t, iv, 32)
		value := NewKSubsetValue(2, iv)
		set, err := value.ToSetEnum()
		if err != nil {
			t.Fatal(err)
		}
		javaFcnSetSize(t, set, 496)
		array := set.Elems.ToArray()
		for _, v := range array {
			if v == nil {
				t.Fatal("expected non-null element")
			}
			javaFcnSetSize(t, v, 2)
		}
	})
	t.Run("testEnumerateN33", func(t *testing.T) {
		iv := NewIntervalValue(1, 33)
		javaFcnSetSize(t, iv, 33)
		value := NewKSubsetValue(2, iv)
		set, err := value.ToSetEnum()
		if err != nil {
			t.Fatal(err)
		}
		javaFcnSetSize(t, set, 528)
		array := set.Elems.ToArray()
		for _, v := range array {
			if v == nil {
				t.Fatal("expected non-null element")
			}
			javaFcnSetSize(t, v, 2)
		}
	})
	t.Run("testEnumerateN63", func(t *testing.T) {
		iv := NewIntervalValue(1, 63)
		javaFcnSetSize(t, iv, 63)
		value := NewKSubsetValue(2, iv)
		set, err := value.ToSetEnum()
		if err != nil {
			t.Fatal(err)
		}
		javaFcnSetSize(t, set, 1953)
		array := set.Elems.ToArray()
		for _, v := range array {
			if v == nil {
				t.Fatal("expected non-null element")
			}
			javaFcnSetSize(t, v, 2)
		}
	})
	t.Run("testEnumerateN64", func(t *testing.T) {
		iv := NewIntervalValue(1, 64)
		javaFcnSetSize(t, iv, 64)
		value := NewKSubsetValue(2, iv)
		set, err := value.ToSetEnum()
		if err != nil {
			t.Fatal(err)
		}
		javaFcnSetSize(t, set, 2016)
		array := set.Elems.ToArray()
		for _, v := range array {
			if v == nil {
				t.Fatal("expected non-null element")
			}
			javaFcnSetSize(t, v, 2)
		}
		_, err = NewKSubsetValue(42, iv).ToSetEnum()
		if _, ok := err.(*IllegalArgumentException); !ok {
			t.Fatalf("expected IllegalArgumentException, got %T: %v", err, err)
		}
	})
	t.Run("testNormalization", func(t *testing.T) {
		vals := NewValueVec(10)
		vals.Add(NewIntValue(1))
		vals.Add(NewIntValue(7))
		vals.Add(NewIntValue(42))
		vals.Add(NewIntValue(42))
		vals.Add(NewIntValue(23))
		set := NewSetEnumValueVec(vals, false)
		value := NewKSubsetValue(2, set)
		javaFcnSetSize(t, value, 6)
		e := value.Elements()
		javaFcnSetEquals(t, NewSetEnumValue([]Value{NewIntValue(1), NewIntValue(7)}, false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue([]Value{NewIntValue(1), NewIntValue(23)}, false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue([]Value{NewIntValue(1), NewIntValue(42)}, false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue([]Value{NewIntValue(7), NewIntValue(23)}, false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue([]Value{NewIntValue(7), NewIntValue(42)}, false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue([]Value{NewIntValue(23), NewIntValue(42)}, false), javaFcnSetNext(t, e))
		javaFcnSetNull(t, e)
	})
	t.Run("testKSubsetFingerprintingS009", func(t *testing.T) { javaKSubsetDoTest(t, []int{1, 9}, NewIntervalValue(1, 9)) })
	t.Run("testKSubsetFingerprintingS032", func(t *testing.T) {
		javaKSubsetDoTest(t, []int{1, 2, 3, 4, 5, 28, 29, 30, 31, 32}, NewIntervalValue(1, 32))
	})
	t.Run("testKSubsetFingerprintingS033", func(t *testing.T) {
		javaKSubsetDoTest(t, []int{1, 2, 3, 4, 5, 29, 30, 31, 32, 33}, NewIntervalValue(1, 33))
	})
	t.Run("testKSubsetFingerprintingS063", func(t *testing.T) { javaKSubsetDoTest(t, []int{1, 2, 3, 4, 60, 61, 62, 63}, NewIntervalValue(1, 63)) })
	t.Run("testInvalidKDenotesEmptySet", func(t *testing.T) {
		base := NewIntervalValue(1, 3)
		for _, k := range []int{-1, 4} {
			value := NewKSubsetValue(k, base)
			javaFcnSetSize(t, value, 0)
			empty, err := IsEmptyValue(value)
			if err != nil || !empty {
				t.Fatalf("isEmpty=%v/%v, want true", empty, err)
			}
			javaFcnSetEquals(t, EmptySet, value)
			javaFcnSetEquals(t, value, EmptySet)
			cmp, err := value.Compare(EmptySet)
			if err != nil || cmp != 0 {
				t.Fatalf("value compare=%d/%v, want 0", cmp, err)
			}
			cmp, err = EmptySet.Compare(value)
			if err != nil || cmp != 0 {
				t.Fatalf("empty compare=%d/%v, want 0", cmp, err)
			}
			expectedHash, actualHash := ValueJavaHashCode(EmptySet), ValueJavaHashCode(value)
			if expectedHash != actualHash {
				t.Fatalf("hash=%d, want %d", actualHash, expectedHash)
			}
			javaFcnSetNull(t, value.Elements())
			javaFcnSetNull(t, javaSubsetRandomOrder(t, value))
			set, err := value.ToSetEnum()
			if err != nil {
				t.Fatal(err)
			}
			javaFcnSetEquals(t, EmptySet, set)
			if actual := value.String(); actual != "{}" {
				t.Fatalf("string=%q, want {}", actual)
			}
		}
	})
	t.Run("testToStringLargeSwallowsCountError", func(t *testing.T) {
		large := NewIntervalValue(1, 64)
		expected := "{s \\in SUBSET (1..64) : Cardinality(s) = 32}"
		if actual := NewKSubsetValue(32, large).String(); actual != expected {
			t.Fatalf("string=%q, want %q", actual, expected)
		}
	})
}
