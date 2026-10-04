/*******************************************************************************
 * Copyright (c) 2018 Microsoft Research. All rights reserved.
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
// Complete original SetOfRcrdValueTest, retaining its full n/m/k matrix.
package tlc

import (
	"fmt"
	"testing"
)

func javaRecordSetNames(n int) []*UniqueString {
	names := make([]*UniqueString, n)
	for i := range names {
		names[i] = UniqueStringOf(fmt.Sprintf("N%d", i))
	}
	return names
}
func javaRecordSetStrings(n int, str string) []Value {
	values := make([]Value, n)
	for i := range values {
		values[i] = NewStringValue(fmt.Sprintf("%s%d", str, i))
	}
	return values
}
func javaRecordSetRanges(n int, names []*UniqueString) []Value {
	values := make([]Value, len(names))
	for i := range values {
		values[i] = NewSetEnumValue(javaRecordSetStrings(n, string(rune(97+i))), false)
	}
	return values
}
func javaRecordSetNew(t *testing.T, names []*UniqueString, values []Value, norm bool) *SetOfRcdsValue {
	t.Helper()
	s, err := NewSetOfRcdsValue(names, values, norm)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func javaRecordSetSize(t *testing.T, s *SetOfRcdsValue) int {
	t.Helper()
	n, err := s.Size()
	if err != nil {
		t.Fatal(err)
	}
	return n
}
func javaRecordSetIndexed(t *testing.T, s *SetOfRcdsValue, k int) *randomIndexedValueEnumeration {
	t.Helper()
	e, err := randomValueEnumeration(s, k)
	if err != nil {
		t.Fatal(err)
	}
	indexed, ok := e.(*randomIndexedValueEnumeration)
	if !ok {
		t.Fatalf("%T, want SubsetEnumerator", e)
	}
	return indexed
}
func javaRecordSetSubset(t *testing.T, s *SetOfRcdsValue, k int) Value {
	t.Helper()
	v, handled, err := randomSubsetOfProductValue(k, s)
	if err != nil || !handled {
		t.Fatalf("getRandomSubset=%v/%v", handled, err)
	}
	return v
}
func javaRecordSetCheckElements(t *testing.T, names []*UniqueString, values []Value, s *SetOfRcdsValue, e *randomIndexedValueEnumeration) {
	t.Helper()
	for i := 0; i < javaRecordSetSize(t, s); i++ {
		rcd := e.element(i).(*RecordValue)
		if len(names) != len(rcd.Names) {
			t.Fatalf("names length=%d, want %d", len(rcd.Names), len(names))
		}
		for j, name := range names {
			if !name.Equal(rcd.Names[j]) {
				t.Fatalf("name[%d]=%v, want %v", j, rcd.Names[j], name)
			}
		}
		for j, v := range rcd.Values {
			member, err := values[j].Member(v)
			if err != nil || !member {
				t.Fatalf("field[%d] member=%v/%v", j, member, err)
			}
		}
	}
}
func javaRecordSetCollect(t *testing.T, names []*UniqueString, s *SetOfRcdsValue, subset Value, hash *javaFcnSetHash) {
	t.Helper()
	e := javaFcnSetElements(t, subset)
	for v := javaFcnSetNext(t, e); v != nil; v = javaFcnSetNext(t, e) {
		rcd := v.(*RecordValue)
		if len(rcd.Names) != len(names) {
			t.Fatalf("names length=%d, want %d", len(rcd.Names), len(names))
		}
		if len(rcd.Values) != len(names) {
			t.Fatalf("values length=%d, want %d", len(rcd.Values), len(names))
		}
		hash.add(t, rcd)
		member, err := s.Member(rcd)
		if err != nil || !member {
			t.Fatalf("member=%v/%v", member, err)
		}
	}
}
func TestJavaSetOfRcrdValue(t *testing.T) {
	oldIntern := internTable
	internTable = NewInternTable(1024)
	oldPoly := FP64IrredPoly()
	oldRng := ResetRandomEnumerableValues()
	oldChecker, oldSimulator := MainChecker(), CurrentSimulator()
	SetMainChecker(nil)
	SetSimulator(nil)
	t.Cleanup(func() {
		internTable = oldIntern
		FP64InitPoly(oldPoly)
		SetRandomEnumerableGenerator(oldRng)
		SetMainChecker(oldChecker)
		SetSimulator(oldSimulator)
	})
	FP64Init() // Original @BeforeClass.
	t.Run("testSimple", func(t *testing.T) {
		names := javaRecordSetNames(3)
		values := []Value{NewSetEnumValue(javaRecordSetStrings(7, "a"), true), NewIntervalValue(1, 2), NewIntervalValue(1, 4)}
		s := javaRecordSetNew(t, names, values, true)
		javaRecordSetCheckElements(t, names, values, s, javaRecordSetIndexed(t, s, javaRecordSetSize(t, s)))
	})
	t.Run("testRangeSubsetValue", func(t *testing.T) {
		names := javaRecordSetNames(4)
		s := javaRecordSetNew(t, names, javaRecordSetRanges(2, names), true)
		actual := &javaFcnSetHash{}
		for _, v := range javaFcnSetAll(t, javaRecordSetIndexed(t, s, javaRecordSetSize(t, s))) {
			actual.add(t, v)
		}
		javaFcnSetHashSize(t, actual, javaRecordSetSize(t, s))
		expected := &javaFcnSetHash{}
		for _, v := range javaFcnSetAll(t, s.Elements()) {
			expected.add(t, v)
		}
		javaFcnSetHashSize(t, actual, len(expected.values))
		for _, v := range actual.values {
			if !expected.contains(t, v) {
				t.Fatal("HashSets differ")
			}
		}
	})
	t.Run("testRandomSubset", func(t *testing.T) {
		names := javaRecordSetNames(4)
		s := javaRecordSetNew(t, names, javaRecordSetRanges(3, names), true)
		size := 27
		subset := javaRecordSetSubset(t, s, size)
		values := &javaFcnSetHash{}
		javaRecordSetCollect(t, names, s, subset, values)
		javaFcnSetHashSize(t, values, size)
	})
	t.Run("testRandomSubsetVaryingParameters", func(t *testing.T) {
		for n := 1; n < 7; n++ {
			names := javaRecordSetNames(n)
			for m := 1; m < 5; m++ {
				s := javaRecordSetNew(t, names, javaRecordSetRanges(m, names), true)
				values := &javaFcnSetHash{}
				for kOutOfN := 0; kOutOfN < javaRecordSetSize(t, s); kOutOfN++ {
					subset := javaRecordSetSubset(t, s, kOutOfN)
					javaRecordSetCollect(t, names, s, subset, values)
					javaFcnSetHashSize(t, values, kOutOfN)
					values.buckets = nil
					values.values = nil // Original HashSet.clear().
				}
			}
		}
	})
	t.Run("testRandomSubsetAstronomically", func(t *testing.T) {
		names := javaRecordSetNames(10)
		s := javaRecordSetNew(t, names, javaRecordSetRanges(50, names), true)
		k := 10000
		subset := javaRecordSetSubset(t, s, k)
		values := &javaFcnSetHash{}
		javaRecordSetCollect(t, names, s, subset, values)
		javaFcnSetHashSize(t, values, k)
	})
	t.Run("testEmptyNonEnumerableField", func(t *testing.T) {
		s := javaRecordSetNew(t, javaRecordSetNames(2), []Value{Nat(), javaFcnSetEmpty()}, false)
		if all := javaFcnSetAll(t, s.Elements()); len(all) != 0 {
			t.Fatalf("all size=%d, want 0", len(all))
		}
	})
	t.Run("testRandomSubsetEmptyNonEnumerableField", func(t *testing.T) {
		s := javaRecordSetNew(t, javaRecordSetNames(2), []Value{Nat(), javaFcnSetEmpty()}, false)
		javaFcnSetSize(t, s, 0)
		big, err := randomProductNeedsBigInteger(s)
		if err != nil || big {
			t.Fatalf("needBigInteger=%v/%v, want false", big, err)
		}
		e, err := randomValueEnumeration(s, 1)
		if err != nil {
			t.Fatal(err)
		}
		if all := javaFcnSetAll(t, e); len(all) != 0 {
			t.Fatalf("elements(1).all size=%d, want 0", len(all))
		}
		javaFcnSetSize(t, javaRecordSetSubset(t, s, 1), 0)
	})
}
