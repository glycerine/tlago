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
// Complete original SetOfFcnsValueTest. Preserve indexed enumeration and the
// complete four-domain/seven-sample large-set matrix.
package tlc

import (
	"fmt"
	"testing"
)

func javaFcnSetEmpty() *SetEnumValue { return NewSetEnumValueVec(NewValueVec(0), true) }
func javaFcnSetModels() []Value {
	return []Value{MakeModelValue("m1"), MakeModelValue("m2"), MakeModelValue("m3")}
}
func javaFcnSetSize(t *testing.T, v Value, want int) {
	t.Helper()
	n, err := v.Size()
	if err != nil || n != want {
		t.Fatalf("size=%d/%v, want %d", n, err, want)
	}
}
func javaFcnSetEquals(t *testing.T, a, b Value) {
	t.Helper()
	if a == nil || b == nil {
		t.Fatal("null value")
	}
	eq, err := a.Equal(b)
	if err != nil || !eq {
		t.Fatalf("equals=%v/%v, want true", eq, err)
	}
}
func javaFcnSetNext(t *testing.T, e ValueEnumeration) Value {
	t.Helper()
	v := e.NextElement()
	if err := e.Err(); err != nil {
		t.Fatal(err)
	}
	return v
}
func javaFcnSetNull(t *testing.T, e ValueEnumeration) {
	t.Helper()
	if v := javaFcnSetNext(t, e); v != nil {
		t.Fatalf("expected null, got %v", v)
	}
}
func javaFcnSetAll(t *testing.T, e ValueEnumeration) []Value {
	t.Helper()
	a := []Value{}
	for v := javaFcnSetNext(t, e); v != nil; v = javaFcnSetNext(t, e) {
		a = append(a, v)
	}
	return a
}
func javaFcnSetMember(t *testing.T, f *SetOfFcnsValue, v Value) {
	t.Helper()
	ok, err := f.Member(v)
	if err != nil || !ok {
		t.Fatalf("member=%v/%v, want true", ok, err)
	}
}
func javaFcnSetSubset(t *testing.T, f *SetOfFcnsValue, k int) Value {
	t.Helper()
	s, handled, err := randomSubsetOfProductValue(k, f)
	if err != nil || !handled {
		t.Fatalf("getRandomSubset=%v/%v", handled, err)
	}
	return s
}
func javaFcnSetElements(t *testing.T, v Value) ValueEnumeration {
	t.Helper()
	e, ok := asEnumerable(v)
	if !ok {
		t.Fatalf("not enumerable: %T", v)
	}
	return e.Elements()
}
func javaFcnSetIndexed(t *testing.T, f *SetOfFcnsValue, k int) *randomIndexedValueEnumeration {
	t.Helper()
	e, err := randomValueEnumeration(f, k)
	if err != nil {
		t.Fatal(err)
	}
	indexed, ok := e.(*randomIndexedValueEnumeration)
	if !ok {
		t.Fatalf("%T, want SubsetEnumerator", e)
	}
	return indexed
}
func javaFcnSetDimensions(t *testing.T, v Value, n int) *FcnRcdValue {
	t.Helper()
	f, ok := v.(*FcnRcdValue)
	if !ok {
		t.Fatalf("%T, want FcnRcdValue", v)
	}
	if len(f.Domain) != n {
		t.Fatalf("domain length=%d, want %d", len(f.Domain), n)
	}
	if len(f.Values) != n {
		t.Fatalf("values length=%d, want %d", len(f.Values), n)
	}
	return f
}

// HashSet with Java Value.hashCode and equals, retaining actual values.
type javaFcnSetHash struct {
	buckets map[int32][]Value
	values  []Value
}

func (s *javaFcnSetHash) contains(t *testing.T, v Value) bool {
	t.Helper()
	for _, old := range s.buckets[ValueJavaHashCode(v)] {
		eq, err := v.Equal(old)
		if err != nil {
			t.Fatal(err)
		}
		if eq {
			return true
		}
	}
	return false
}
func (s *javaFcnSetHash) add(t *testing.T, v Value) {
	t.Helper()
	// HashSet.add/HashMap.put hashes its incoming key once, including when
	// inserting a new value. Keep the same fingerprint call count here.
	h := ValueJavaHashCode(v)
	for _, old := range s.buckets[h] {
		eq, err := v.Equal(old)
		if err != nil {
			t.Fatal(err)
		}
		if eq {
			return
		}
	}
	if s.buckets == nil {
		s.buckets = map[int32][]Value{}
	}
	s.buckets[h] = append(s.buckets[h], v)
	s.values = append(s.values, v)
}
func javaFcnSetHashSize(t *testing.T, s *javaFcnSetHash, want int) {
	t.Helper()
	if len(s.values) != want {
		t.Fatalf("HashSet size=%d, want %d", len(s.values), want)
	}
}
func TestJavaSetOfFcnsValue(t *testing.T) {
	oldIntern := internTable
	internTable = NewInternTable(1024)
	modelValues.Lock()
	oldCount, oldTable, oldMVs := modelValues.count, modelValues.table, modelValues.mvs
	modelValues.Unlock()
	ModelValueInit()
	oldPoly := FP64IrredPoly()
	oldRng := ResetRandomEnumerableValues()
	oldChecker, oldSimulator := MainChecker(), CurrentSimulator()
	SetMainChecker(nil)
	SetSimulator(nil)
	t.Cleanup(func() {
		internTable = oldIntern
		modelValues.Lock()
		modelValues.count, modelValues.table, modelValues.mvs = oldCount, oldTable, oldMVs
		modelValues.Unlock()
		FP64InitPoly(oldPoly)
		SetRandomEnumerableGenerator(oldRng)
		SetMainChecker(oldChecker)
		SetSimulator(oldSimulator)
	})
	t.Run("testRangeSubsetValue", func(t *testing.T) {
		values := javaFcnSetModels()
		f := NewSetOfFcnsValue(NewSetEnumValue(values, true), NewSubsetValue(NewSetEnumValue(stringValues("a", "b", "c"), true)))
		javaFcnSetSize(t, f, 512)
		e := javaFcnSetIndexed(t, f, 512)
		empty := javaFcnSetEmpty()
		expected := []Value{empty, NewSetEnumValue(stringValues("a"), true), NewSetEnumValue(stringValues("b"), true), NewSetEnumValue(stringValues("c"), true), NewSetEnumValue(stringValues("a", "b"), true)}
		for i, last := range expected {
			javaFcnSetEquals(t, NewFcnRcdValue(values, []Value{empty, empty, last}, true), e.element(i))
		}
		full := NewSetEnumValue(stringValues("a", "b", "c"), true)
		javaFcnSetEquals(t, NewFcnRcdValue(values, []Value{full, full, full}, true), e.element(511))
	})
	t.Run("testDomainEmpty", func(t *testing.T) {
		domain := javaFcnSetEmpty()
		f := NewSetOfFcnsValue(domain, NewSetEnumValue(stringValues("a", "b", "c"), true))
		javaFcnSetSize(t, f, 1)
		e := f.Elements()
		javaFcnSetEquals(t, NewFcnRcdValue([]Value{}, []Value{}, true), javaFcnSetNext(t, e))
		javaFcnSetNull(t, e)
		subset := javaFcnSetSubset(t, f, 5)
		se := javaFcnSetElements(t, subset)
		javaFcnSetSize(t, subset, 1)
		javaFcnSetEquals(t, NewFcnRcdValue([]Value{}, []Value{}, true), javaFcnSetNext(t, se))
		javaFcnSetNull(t, se)
	})
	t.Run("testRangeEmpty", func(t *testing.T) {
		f := NewSetOfFcnsValue(NewIntervalValue(1, 2), NewSetEnumValueVec(NewValueVec(10), true))
		javaFcnSetSize(t, f, 0)
		javaFcnSetNull(t, f.Elements())
		s := javaFcnSetSubset(t, f, 5)
		javaFcnSetSize(t, s, 0)
		javaFcnSetNull(t, javaFcnSetElements(t, s))
		javaFcnSetEquals(t, javaFcnSetEmpty(), s)
	})
	t.Run("testDomainAndRangeEmpty", func(t *testing.T) {
		domain := javaFcnSetEmpty()
		f := NewSetOfFcnsValue(domain, javaFcnSetEmpty())
		javaFcnSetSize(t, f, 1)
		e := f.Elements()
		javaFcnSetEquals(t, NewFcnRcdValue([]Value{}, []Value{}, true), javaFcnSetNext(t, e))
		javaFcnSetNull(t, e)
		subset := javaFcnSetSubset(t, f, 5)
		se := javaFcnSetElements(t, subset)
		javaFcnSetSize(t, subset, 1)
		javaFcnSetEquals(t, NewFcnRcdValue([]Value{}, []Value{}, true), javaFcnSetNext(t, se))
		javaFcnSetNull(t, se)
	})
	t.Run("testRandomSubsetAndValueEnumerator", func(t *testing.T) {
		domain := javaFcnSetModels()
		f := NewSetOfFcnsValue(NewSetEnumValue(domain, true), NewSetEnumValue(stringValues("a", "b", "c"), true))
		javaFcnSetSize(t, f, 27)
		FP64Init()
		a := &javaFcnSetHash{}
		e := javaFcnSetIndexed(t, f, 27)
		for i := 0; i < 27; i++ {
			v := javaFcnSetDimensions(t, e.element(i), 3)
			a.add(t, v)
		}
		subset := javaFcnSetSubset(t, f, 27)
		b := &javaFcnSetHash{}
		se := javaFcnSetElements(t, subset)
		for v := javaFcnSetNext(t, se); v != nil; v = javaFcnSetNext(t, se) {
			fcn := javaFcnSetDimensions(t, v, 3)
			b.add(t, fcn)
			javaFcnSetMember(t, f, fcn)
		}
		javaFcnSetHashSize(t, b, len(a.values))
		if len(a.values) != len(b.values) {
			t.Fatal("HashSets differ")
		}
		for _, v := range a.values {
			if !b.contains(t, v) {
				t.Fatal("HashSets differ")
			}
		}
	})
	t.Run("testDomainModelValue", func(t *testing.T) {
		domain := javaFcnSetModels()
		f := NewSetOfFcnsValue(NewSetEnumValue(domain, true), NewSetEnumValue(stringValues("a", "b", "c"), true))
		javaFcnSetSize(t, f, 27)
		FP64Init()
		set := &javaFcnSetHash{}
		e := javaFcnSetIndexed(t, f, 27)
		for i := 0; i < 27; i++ {
			v := javaFcnSetDimensions(t, e.element(i), 3)
			set.add(t, v)
			javaFcnSetMember(t, f, v)
		}
		javaFcnSetHashSize(t, set, 27)
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("a", "a", "a"), true), e.element(0))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("a", "a", "b"), true), e.element(1))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("a", "a", "c"), true), e.element(2))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("a", "b", "a"), true), e.element(3))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("a", "b", "b"), true), e.element(4))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("a", "b", "c"), true), e.element(5))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("a", "c", "a"), true), e.element(6))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("a", "c", "b"), true), e.element(7))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("a", "c", "c"), true), e.element(8))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("b", "a", "a"), true), e.element(9))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("b", "a", "b"), true), e.element(10))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("b", "a", "c"), true), e.element(11))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("b", "b", "a"), true), e.element(12))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("b", "b", "b"), true), e.element(13))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("b", "b", "c"), true), e.element(14))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("b", "c", "a"), true), e.element(15))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("b", "c", "b"), true), e.element(16))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("b", "c", "c"), true), e.element(17))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("c", "a", "a"), true), e.element(18))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("c", "a", "b"), true), e.element(19))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("c", "a", "c"), true), e.element(20))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("c", "b", "a"), true), e.element(21))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("c", "b", "b"), true), e.element(22))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("c", "b", "c"), true), e.element(23))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("c", "c", "a"), true), e.element(24))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("c", "c", "b"), true), e.element(25))
		javaFcnSetEquals(t, NewFcnRcdValue(domain, stringValues("c", "c", "c"), true), e.element(26))
	})
	t.Run("testDomainIntervalRangeSetEnumValueSize9", func(t *testing.T) {
		domain := NewIntervalValue(1, 2)
		f := NewSetOfFcnsValue(domain, NewSetEnumValue(stringValues("a", "b", "c"), true))
		javaFcnSetSize(t, f, 9)
		e := javaFcnSetIndexed(t, f, 9)
		for i := 0; i < 9; i++ {
			v := javaFcnSetDimensions(t, e.element(i), 2)
			javaFcnSetMember(t, f, v)
		}
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "a")), e.element(0))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "b")), e.element(1))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "c")), e.element(2))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("b", "a")), e.element(3))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("b", "b")), e.element(4))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("b", "c")), e.element(5))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("c", "a")), e.element(6))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("c", "b")), e.element(7))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("c", "c")), e.element(8))
	})
	t.Run("testDomainIntervalRangeSetEnumValueSize27", func(t *testing.T) {
		domain := NewIntervalValue(1, 3)
		f := NewSetOfFcnsValue(domain, NewSetEnumValue(stringValues("a", "b", "c"), true))
		javaFcnSetSize(t, f, 27)
		e := javaFcnSetIndexed(t, f, 27)
		for i := 0; i < 27; i++ {
			v := javaFcnSetDimensions(t, e.element(i), 3)
			javaFcnSetMember(t, f, v)
		}
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "a", "a")), e.element(0))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "a", "b")), e.element(1))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "a", "c")), e.element(2))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "b", "a")), e.element(3))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "b", "b")), e.element(4))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "b", "c")), e.element(5))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "c", "a")), e.element(6))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "c", "b")), e.element(7))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "c", "c")), e.element(8))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("b", "a", "a")), e.element(9))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("b", "a", "b")), e.element(10))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("b", "a", "c")), e.element(11))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("b", "b", "a")), e.element(12))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("b", "b", "b")), e.element(13))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("b", "b", "c")), e.element(14))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("b", "c", "a")), e.element(15))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("b", "c", "b")), e.element(16))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("b", "c", "c")), e.element(17))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("c", "a", "a")), e.element(18))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("c", "a", "b")), e.element(19))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("c", "a", "c")), e.element(20))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("c", "b", "a")), e.element(21))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("c", "b", "b")), e.element(22))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("c", "b", "c")), e.element(23))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("c", "c", "a")), e.element(24))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("c", "c", "b")), e.element(25))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("c", "c", "c")), e.element(26))
	})
	t.Run("testDomainIntervalRangeSetEnumValueSize256", func(t *testing.T) {
		domain := NewIntervalValue(1, 4)
		f := NewSetOfFcnsValue(domain, NewSetEnumValue(stringValues("a", "b", "c", "d"), true))
		javaFcnSetSize(t, f, 256)
		e := javaFcnSetIndexed(t, f, 256)
		for i := 0; i < 256; i++ {
			v := javaFcnSetDimensions(t, e.element(i), 4)
			javaFcnSetMember(t, f, v)
		}
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "a", "a", "a")), e.element(0))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "a", "a", "b")), e.element(1))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "a", "a", "c")), e.element(2))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("a", "a", "a", "d")), e.element(3))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("d", "d", "d", "b")), e.element(253))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("d", "d", "d", "c")), e.element(254))
		javaFcnSetEquals(t, NewFcnRcdIntervalValue(domain, stringValues("d", "d", "d", "d")), e.element(255))
	})

	t.Run("testRandomSubsetFromReallyLarge", func(t *testing.T) {
		sets := []*SetOfFcnsValue{
			NewSetOfFcnsValue(NewIntervalValue(1, 11), NewSetEnumValue(stringValues("a", "b", "c", "d", "e", "f", "g", "h", "i", "j"), true)),
			NewSetOfFcnsValue(NewIntervalValue(1, 44), NewIntervalValue(1, 20)),
			NewSetOfFcnsValue(NewIntervalValue(1, 121), NewIntervalValue(1, 19)),
			NewSetOfFcnsValue(NewIntervalValue(1, 321), NewIntervalValue(1, 29)),
		}
		for i, f := range sets {
			if _, err := f.Size(); javaRuntimeException(err) == nil {
				t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
			}
			for _, k := range []int{0, 1, 2, 799, 1024, 8932, 16933} {
				t.Run(fmt.Sprintf("domain%d/sample%d", i, k), func(t *testing.T) {
					subset := javaFcnSetSubset(t, f, k)
					javaFcnSetSize(t, subset, k)
					FP64Init()
					set := &javaFcnSetHash{}
					e := javaFcnSetElements(t, subset)
					for v := javaFcnSetNext(t, e); v != nil; v = javaFcnSetNext(t, e) {
						fcn := v.(*FcnRcdValue)
						javaFcnSetMember(t, f, fcn)
						set.add(t, fcn)
					}
					javaFcnSetHashSize(t, set, k)
				})
			}
		}
	})
	t.Run("testEmptyNonEnumerableDomain", func(t *testing.T) {
		f := NewSetOfFcnsValue(Nat(), javaFcnSetEmpty())
		all := javaFcnSetAll(t, f.Elements())
		if len(all) != 0 {
			t.Fatalf("all().size=%d, want 0", len(all))
		}
	})
	t.Run("testUnitNonEnumerableRange", func(t *testing.T) {
		f := NewSetOfFcnsValue(javaFcnSetEmpty(), Nat())
		all := javaFcnSetAll(t, f.Elements())
		if len(all) != 1 {
			t.Fatalf("list size=%d, want 1", len(all))
		}
		javaFcnSetEquals(t, EmptyFcn, all[0])
	})
	t.Run("testUnitNonEnumerableRangeInterval", func(t *testing.T) {
		f := NewSetOfFcnsValue(NewIntervalValue(1, 0), Nat())
		all := javaFcnSetAll(t, f.Elements())
		if len(all) != 1 {
			t.Fatalf("list size=%d, want 1", len(all))
		}
		javaFcnSetEquals(t, EmptyFcn, all[0])
	})
	for _, tc := range []struct {
		name   string
		domain Value
	}{
		{"testNonEnumerableRange", NewSetEnumValue(stringValues("d1"), true)},
		{"testNonEnumerableRangeInterval", NewIntervalValue(1, 2)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := NewSetOfFcnsValue(tc.domain, Nat())
			e := f.Elements()
			err := e.Err()
			failure := javaRuntimeException(err)
			if failure == nil {
				t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
			}
			expected := "Attempted to enumerate a set of the form [D -> R],but the range R:\nNat\ncannot be enumerated."
			if failure.Error() != expected {
				t.Fatalf("message=%q, want %q", failure.Error(), expected)
			}
		})
	}
	t.Run("testRandomSubsetEmptyNonEnumerableDomain", func(t *testing.T) {
		f := NewSetOfFcnsValue(Nat(), javaFcnSetEmpty())
		javaFcnSetSize(t, f, 0)
		big, err := randomProductNeedsBigInteger(f)
		if err != nil || big {
			t.Fatalf("needBigInteger=%v/%v, want false", big, err)
		}
		e, err := randomValueEnumeration(f, 1)
		if err != nil {
			t.Fatal(err)
		}
		if all := javaFcnSetAll(t, e); len(all) != 0 {
			t.Fatalf("elements(1).all().size=%d, want 0", len(all))
		}
		javaFcnSetSize(t, javaFcnSetSubset(t, f, 1), 0)
	})
}
