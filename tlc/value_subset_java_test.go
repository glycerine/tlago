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
// Complete original SubsetValueTest. Fixture shuffle has a separate default
// java.util.Random; the TLC subset generator retains source seed 15041980.
package tlc

import (
	"testing"
)

func javaSubsetFixture(rng *JavaRandom, ss ...string) []Value {
	values := stringValues(ss...)
	// Collections.shuffle(RandomAccess List): descending Fisher-Yates.
	for i := len(values); i > 1; i-- {
		j := int(rng.NextIntN(int32(i)))
		values[i-1], values[j] = values[j], values[i-1]
	}
	return values
}
func javaSubsetNum(t *testing.T, s interface{ NumberOfKElements(int) (int64, error) }, k int) int64 {
	t.Helper()
	n, err := s.NumberOfKElements(k)
	if err != nil {
		t.Fatal(err)
	}
	return n
}
func javaSubsetNumEq(t *testing.T, s interface{ NumberOfKElements(int) (int64, error) }, k int, want int64) {
	t.Helper()
	if n := javaSubsetNum(t, s, k); n != want {
		t.Fatalf("numberOfKElements(%d)=%d, want %d", k, n, want)
	}
}
func javaSubsetRandom(t *testing.T, s *SubsetValue, k int) ValueEnumeration {
	t.Helper()
	e, err := randomValueEnumeration(s, k)
	if err != nil {
		t.Fatal(err)
	}
	return e
}
func javaSubsetRandomOrder(t *testing.T, s *KSubsetValue) ValueEnumeration {
	t.Helper()
	e, err := randomizedValueEnumeration(s)
	if err != nil {
		t.Fatal(err)
	}
	return e
}
func javaSubsetIndexed(t *testing.T, e ValueEnumeration) *randomIndexedValueEnumeration {
	t.Helper()
	x, ok := e.(*randomIndexedValueEnumeration)
	if !ok {
		t.Fatalf("%T, want SubsetEnumerator", e)
	}
	return x
}
func javaSubsetTossing(t *testing.T, e ValueEnumeration) *coinTossingSubsetEnumeration {
	t.Helper()
	x, ok := e.(*coinTossingSubsetEnumeration)
	if !ok {
		t.Fatalf("%T, want CoinTossingSubsetEnumerator", e)
	}
	return x
}
func javaSubsetBounds(t *testing.T, v Value, inner Value) {
	t.Helper()
	n := v.(*SetEnumValue).Elems.Len()
	if n < 0 || n > javaSubsetEnumerationSize(t, inner) {
		t.Fatalf("element size %d outside source bounds", n)
	}
}
func javaSubsetSortedK(t *testing.T, s *SubsetValue, k int) ValueEnumeration {
	t.Helper()
	e := s.KElements(k)
	if err := e.Err(); err != nil {
		t.Fatal(err)
	}
	x, ok := e.(*KElementEnumeration)
	if !ok {
		t.Fatalf("%T, want KElementEnumerator", e)
	}
	x, err := x.Sort()
	if err != nil {
		t.Fatal(err)
	}
	return x
}
func javaSubsetAllK(t *testing.T, s *SubsetValue) []Value {
	t.Helper()
	values := make([]Value, 0, javaSubsetEnumerationSize(t, s))
	for k := 0; k <= javaSubsetEnumerationSize(t, s.Set); k++ {
		values = append(values, javaFcnSetAll(t, s.KElements(k))...)
	}
	return values
}
func javaSubsetNormalized(t *testing.T, s *SubsetValue, values []Value) {
	t.Helper()
	enumerated := NewSetEnumValueVec(NewValueVecFrom(values), true)
	set, err := s.ToSetEnum()
	if err != nil {
		t.Fatal(err)
	}
	javaFcnSetEquals(t, set.Normalize(), enumerated)
}

// TreeSet's comparator uses Value.compareTo, including first-key validation.
// Keep its red/black insertion and rotations so comparator evaluation order is
// the same as Java TreeMap.put with an explicit comparator.
type javaSubsetTreeNode struct {
	value               Value
	left, right, parent *javaSubsetTreeNode
	red                 bool
}
type javaSubsetTree struct {
	root *javaSubsetTreeNode
	size int
}

func javaSubsetRed(n *javaSubsetTreeNode) bool { return n != nil && n.red }
func (s *javaSubsetTree) rotateLeft(p *javaSubsetTreeNode) {
	r := p.right
	p.right = r.left
	if r.left != nil {
		r.left.parent = p
	}
	r.parent = p.parent
	if p.parent == nil {
		s.root = r
	} else if p.parent.left == p {
		p.parent.left = r
	} else {
		p.parent.right = r
	}
	r.left = p
	p.parent = r
}
func (s *javaSubsetTree) rotateRight(p *javaSubsetTreeNode) {
	l := p.left
	p.left = l.right
	if l.right != nil {
		l.right.parent = p
	}
	l.parent = p.parent
	if p.parent == nil {
		s.root = l
	} else if p.parent.right == p {
		p.parent.right = l
	} else {
		p.parent.left = l
	}
	l.right = p
	p.parent = l
}
func (s *javaSubsetTree) add(t *testing.T, v Value) {
	t.Helper()
	if s.root == nil {
		if _, err := v.Compare(v); err != nil {
			t.Fatal(err)
		}
		s.root = &javaSubsetTreeNode{value: v}
		s.size = 1
		return
	}
	p := s.root
	for {
		cmp, err := v.Compare(p.value)
		if err != nil {
			t.Fatal(err)
		}
		if cmp == 0 {
			return
		}
		if cmp < 0 {
			if p.left != nil {
				p = p.left
				continue
			}
			p.left = &javaSubsetTreeNode{value: v, parent: p, red: true}
			p = p.left
			break
		}
		if p.right != nil {
			p = p.right
			continue
		}
		p.right = &javaSubsetTreeNode{value: v, parent: p, red: true}
		p = p.right
		break
	}
	s.size++
	for p != s.root && p.parent.red {
		if p.parent == p.parent.parent.left {
			y := p.parent.parent.right
			if javaSubsetRed(y) {
				p.parent.red = false
				y.red = false
				p.parent.parent.red = true
				p = p.parent.parent
			} else {
				if p == p.parent.right {
					p = p.parent
					s.rotateLeft(p)
				}
				p.parent.red = false
				p.parent.parent.red = true
				s.rotateRight(p.parent.parent)
			}
		} else {
			y := p.parent.parent.left
			if javaSubsetRed(y) {
				p.parent.red = false
				y.red = false
				p.parent.parent.red = true
				p = p.parent.parent
			} else {
				if p == p.parent.left {
					p = p.parent
					s.rotateRight(p)
				}
				p.parent.red = false
				p.parent.parent.red = true
				s.rotateLeft(p.parent.parent)
			}
		}
	}
	s.root.red = false
}
func javaSubsetDoTest(t *testing.T, expectedSize int, inner Value, expectedElements int) {
	t.Helper()
	s := NewSubsetValue(inner)
	javaFcnSetSize(t, s, expectedSize)
	unique := &javaSubsetTree{}
	e := javaSubsetRandom(t, s, expectedElements)
	javaSubsetIndexed(t, e)
	for next := javaFcnSetNext(t, e); next != nil; next = javaFcnSetNext(t, e) {
		javaSubsetBounds(t, next, inner)
		unique.add(t, next)
	}
	if unique.size != expectedElements {
		t.Fatalf("TreeSet size=%d, want %d", unique.size, expectedElements)
	}
}
func TestJavaSubsetValue(t *testing.T) {
	oldIntern := internTable
	internTable = NewInternTable(1024)
	oldPoly := FP64IrredPoly()
	oldSeed := RandomEnumerableSeed()
	oldRng := ResetRandomEnumerableValues()
	oldChecker, oldSimulator := MainChecker(), CurrentSimulator()
	SetMainChecker(nil)
	SetSimulator(nil)
	t.Cleanup(func() {
		internTable = oldIntern
		FP64InitPoly(oldPoly)
		SetRandomEnumerableSeed(oldSeed)
		SetRandomEnumerableGenerator(oldRng)
		SetMainChecker(oldChecker)
		SetSimulator(oldSimulator)
	})
	// Original @BeforeClass.
	SetRandomEnumerableSeed(15041980)
	FP64Init()
	shuffle := NewJavaRandomDefault()
	// JUnit runs these methods in one thread, preserving its generator state.
	classRandom := RandomEnumerableGenerator()
	// JUnit MethodSorters.DEFAULT orders by signed String.hashCode, then name.
	t.Run("testKSubsetEnumeratorGTCapacity", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d"), true)
		subset := NewSubsetValue(inner)
		javaFcnSetNull(t, subset.KElements(javaSubsetEnumerationSize(t, inner)+1))
	})
	t.Run("testEmptyEnumerationsAreIndependent", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		subset := NewSubsetValue(EmptySet)
		first, second := subset.Elements(), subset.Elements()
		javaFcnSetEquals(t, EmptySet, javaFcnSetNext(t, first))
		javaFcnSetEquals(t, EmptySet, javaFcnSetNext(t, second))
		javaFcnSetNull(t, first)
		javaFcnSetNull(t, second)
	})
	t.Run("testUnrankKSubsets", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d", "e"), true)
		subset := NewSubsetValue(inner)
		sizeS := javaSubsetEnumerationSize(t, inner)
		for k := 0; k < sizeS; k++ {
			unrank, err := subset.GetUnrank(k)
			if err != nil {
				t.Fatal(err)
			}
			n := javaSubsetNum(t, subset, k)
			unique := &javaFcnSetHash{}
			for i := int64(0); i < n; i++ {
				v := unrank.SubsetAt(i)
				javaFcnSetSize(t, v, k)
				unique.add(t, v)
			}
			javaFcnSetHashSize(t, unique, int(n))
		}
	})
	t.Run("testElementsNormalizedIsNormalized", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		subset := NewSubsetValue(NewIntervalValue(1, 6))
		javaSubsetNormalized(t, subset, javaFcnSetAll(t, subset.ElementsNormalized()))
	})
	t.Run("testUnrank16viaRank", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewIntervalValue(1, 16)
		subset := NewSubsetValue(inner)
		size := javaSubsetEnumerationSize(t, inner)
		sizeS := int64(1) << uint(size)
		unique := &javaFcnSetHash{}
		for k := 0; k <= size; k++ {
			unrank, err := subset.GetUnrank(k)
			if err != nil {
				t.Fatal(err)
			}
			n := javaSubsetNum(t, subset, k)
			for i := int64(0); i < n; i++ {
				v := unrank.SubsetAt(i)
				javaFcnSetSize(t, v, k)
				unique.add(t, v)
			}
		}
		javaFcnSetHashSize(t, unique, int(sizeS))
	})
	t.Run("testRandomSubsetSubset256", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c"), true)
		innerSubset := NewSubsetValue(inner)
		subset := NewSubsetValue(innerSubset)
		expected := int(javaIntOneLeftShift(javaSubsetEnumerationSize(t, innerSubset)))
		javaFcnSetSize(t, subset, expected)
		unique := &javaFcnSetHash{}
		e := javaSubsetRandom(t, subset, expected)
		for next := javaFcnSetNext(t, e); next != nil; next = javaFcnSetNext(t, e) {
			unique.add(t, next)
		}
		javaFcnSetHashSize(t, unique, expected)
	})
	t.Run("testKSubsetEnumerator", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d"), true)
		subset := NewSubsetValue(inner)
		javaSubsetNumEq(t, subset, 0, 1)
		javaSubsetNumEq(t, subset, 1, 4)
		javaSubsetNumEq(t, subset, 2, 6)
		javaSubsetNumEq(t, subset, 3, 4)
		javaSubsetNumEq(t, subset, 4, 1)
		e := subset.KElements(0)
		javaFcnSetEquals(t, javaFcnSetEmpty(), javaFcnSetNext(t, e))
		javaFcnSetNull(t, e)
		e = javaSubsetSortedK(t, subset, 1)
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "a"), false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "b"), false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "c"), false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "d"), false), javaFcnSetNext(t, e))
		javaFcnSetNull(t, e)
		e = javaSubsetSortedK(t, subset, 2)
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b"), false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "a", "c"), false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "a", "d"), false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "b", "c"), false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "b", "d"), false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "c", "d"), false), javaFcnSetNext(t, e))
		javaFcnSetNull(t, e)
		e = javaSubsetSortedK(t, subset, 3)
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c"), false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "d"), false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "a", "c", "d"), false), javaFcnSetNext(t, e))
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "b", "c", "d"), false), javaFcnSetNext(t, e))
		javaFcnSetNull(t, e)
		e = javaSubsetSortedK(t, subset, 4)
		javaFcnSetEquals(t, NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d"), false), javaFcnSetNext(t, e))
		javaFcnSetNull(t, e)
	})
	t.Run("testRandomSubsetE17F1ENeg3", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewIntervalValue(1, 17)
		subset := NewSubsetValue(inner)
		e := javaSubsetRandom(t, subset, 4223)
		javaSubsetIndexed(t, e)
		unique := &javaFcnSetHash{}
		for next := javaFcnSetNext(t, e); next != nil; next = javaFcnSetNext(t, e) {
			javaSubsetBounds(t, next, inner)
			unique.add(t, next)
		}
		javaFcnSetHashSize(t, unique, javaSubsetIndexed(t, e).indices.k)
	})
	t.Run("testRandomSetOfSubsets300", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewIntervalValue(1, 300)
		subset := NewSubsetValue(inner)
		maxLength := 9
		k := 23071
		set, err := subset.GetRandomSetOfSubsetsUpTo(k, maxLength)
		if err != nil {
			t.Fatal(err)
		}
		javaFcnSetSize(t, set, k)
		e := set.Elements()
		for v := javaFcnSetNext(t, e); v != nil; v = javaFcnSetNext(t, e) {
			if size := javaSubsetEnumerationSize(t, v); size > maxLength {
				t.Fatalf("subset size=%d exceeds %d", size, maxLength)
			}
		}
		set.Normalize()
		javaFcnSetSize(t, set, k)
	})
	t.Run("testRandomSetOfSubsets400", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewIntervalValue(1, 400)
		subset := NewSubsetValue(inner)
		maxLength := 9
		k := 23077
		set, err := subset.GetRandomSetOfSubsetsUpTo(k, maxLength)
		if err != nil {
			t.Fatal(err)
		}
		javaFcnSetSize(t, set, k)
		e := set.Elements()
		for v := javaFcnSetNext(t, e); v != nil; v = javaFcnSetNext(t, e) {
			if size := javaSubsetEnumerationSize(t, v); size > maxLength {
				t.Fatalf("subset size=%d exceeds %d", size, maxLength)
			}
		}
		set.Normalize()
		javaFcnSetSize(t, set, k)
	})
	t.Run("testRandomSubsetSubsetNoOverflow", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d", "e"), true)
		innerSubset := NewSubsetValue(inner)
		subset := NewSubsetValue(innerSubset)
		_, err := subset.Size()
		// Source is catch-only: no fail() if size() does not throw.
		if err == nil {
			return
		}
		if javaRuntimeException(err) == nil {
			t.Fatalf("uncaught %T: %v", err, err)
		}
		unique := &javaFcnSetHash{}
		e := javaSubsetRandom(t, subset, 2148)
		javaSubsetTossing(t, e)
		for next := javaFcnSetNext(t, e); next != nil; next = javaFcnSetNext(t, e) {
			unique.add(t, next)
		}
		javaFcnSetHashSize(t, unique, 2148)
	})
	t.Run("testRandomSubsetGeneratorN100", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		unique := &javaFcnSetHash{}
		subset := NewKSubsetValue(3, NewIntervalValue(1, 100))
		e := javaSubsetRandomOrder(t, subset)
		for i := 0; i < 100; i++ {
			unique.add(t, javaFcnSetNext(t, e))
		}
		javaFcnSetHashSize(t, unique, 100)
	})
	t.Run("testNumKSubsetKGTN", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d", "e"), true)
		subset := NewSubsetValue(inner)
		javaSubsetNumEq(t, subset, javaSubsetEnumerationSize(t, inner)+1, 0)
	})
	t.Run("testNumKSubset", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d", "e"), true)
		subset := NewSubsetValue(inner)
		javaSubsetNumEq(t, subset, 0, 1)
		javaSubsetNumEq(t, subset, 1, 5)
		javaSubsetNumEq(t, subset, 2, 10)
		javaSubsetNumEq(t, subset, 3, 10)
		javaSubsetNumEq(t, subset, 4, 5)
		javaSubsetNumEq(t, subset, 5, 1)
	})
	t.Run("testRandomSubsetE32F1ENeg6", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewIntervalValue(1, 32)
		subset := NewSubsetValue(inner)
		e := javaSubsetRandom(t, subset, 2342)
		javaSubsetTossing(t, e)
		unique := &javaFcnSetHash{}
		for next := javaFcnSetNext(t, e); next != nil; next = javaFcnSetNext(t, e) {
			javaSubsetBounds(t, next, inner)
			unique.add(t, next)
		}
		tossing := javaSubsetTossing(t, e)
		if len(unique.values) < tossing.k-100 || len(unique.values) > tossing.k {
			t.Fatalf("unique size=%d outside source bounds", len(unique.values))
		}
	})
	t.Run("testRandomSubsetSubset16", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b"), true)
		innerSubset := NewSubsetValue(inner)
		subset := NewSubsetValue(innerSubset)
		expected := int(javaIntOneLeftShift(javaSubsetEnumerationSize(t, innerSubset)))
		javaFcnSetSize(t, subset, expected)
		unique := &javaFcnSetHash{}
		e := javaSubsetRandom(t, subset, expected)
		for next := javaFcnSetNext(t, e); next != nil; next = javaFcnSetNext(t, e) {
			unique.add(t, next)
		}
		javaFcnSetHashSize(t, unique, expected)
	})
	t.Run("testRandomSubsetGeneratorKNplus1", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		javaFcnSetNull(t, javaSubsetRandomOrder(t, NewKSubsetValue(3, NewIntervalValue(1, 2))))
	})
	t.Run("testRandomSubsetGeneratorN10", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		unique := &javaFcnSetHash{}
		subset := NewKSubsetValue(3, NewIntervalValue(1, 10))
		e := javaSubsetRandomOrder(t, subset)
		for i := 0; i < 1000; i++ {
			unique.add(t, javaFcnSetNext(t, e))
		}
		javaFcnSetHashSize(t, unique, int(javaSubsetNum(t, subset, 3)))
	})
	t.Run("testNumKSubset2", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d", "e", "f", "g", "h"), true)
		subset := NewSubsetValue(inner)
		sum := int32(0)
		for i := 0; i <= javaSubsetEnumerationSize(t, inner); i++ {
			sum = int32(int64(sum) + javaSubsetNum(t, subset, i))
		}
		if want := javaIntOneLeftShift(javaSubsetEnumerationSize(t, inner)); sum != want {
			t.Fatalf("sum=%d, want %d", sum, want)
		}
	})
	t.Run("testNumKSubsetNeg", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d", "e"), true)
		subset := NewSubsetValue(inner)
		javaSubsetNumEq(t, subset, -1, 0)
	})
	t.Run("testNumKSubsetUpTo62", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		for i := 1; i < 62; i++ {
			subset := NewSubsetValue(NewIntervalValue(1, int32(i)))
			sum := int64(0)
			for j := 0; j <= i; j++ {
				sum += javaSubsetNum(t, subset, j)
			}
			if want := int64(1) << uint(i); sum != want {
				t.Fatalf("sum=%d, want %d", sum, want)
			}
		}
	})
	t.Run("testKElementsAreNormalized", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		subset := NewSubsetValue(NewIntervalValue(1, 6))
		javaSubsetNormalized(t, subset, javaSubsetAllK(t, subset))
	})
	t.Run("testRandomSubsetGeneratorK0", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		unique := &javaFcnSetHash{}
		subset := NewKSubsetValue(0, NewIntervalValue(1, 10))
		e := javaSubsetRandomOrder(t, subset)
		for i := 0; i < 100; i++ {
			unique.add(t, javaFcnSetNext(t, e))
		}
		javaFcnSetHashSize(t, unique, 1)
	})
	t.Run("testRandomSetOfSubsets", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewIntervalValue(1, 22)
		subset := NewSubsetValue(inner)
		maxLength := 10
		k := 23131
		set, err := subset.GetRandomSetOfSubsetsUpTo(k, maxLength)
		if err != nil {
			t.Fatal(err)
		}
		javaFcnSetSize(t, set, k)
		e := set.Elements()
		for v := javaFcnSetNext(t, e); v != nil; v = javaFcnSetNext(t, e) {
			if size := javaSubsetEnumerationSize(t, v); size > maxLength {
				t.Fatalf("subset size=%d exceeds %d", size, maxLength)
			}
		}
		set.Normalize()
		javaFcnSetSize(t, set, k)
	})
	t.Run("testKSubsetEnumeratorNegative", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d"), true)
		subset := NewSubsetValue(inner)
		javaFcnSetNull(t, subset.KElements(-1))
	})
	t.Run("testRandomSubsetE5F025", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d", "e"), true)
		size := int(javaIntOneLeftShift(javaSubsetEnumerationSize(t, inner)))
		javaSubsetDoTest(t, size, inner, 8)
	})
	t.Run("testRandomSubsetE5F075", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d", "e"), true)
		size := int(javaIntOneLeftShift(javaSubsetEnumerationSize(t, inner)))
		javaSubsetDoTest(t, size, inner, 24)
	})
	t.Run("testNumKSubsetPreventsOverflow", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewIntervalValue(1, 64)
		subset := NewSubsetValue(inner)
		javaSubsetNumEq(t, subset, 0, 1)
		javaSubsetNumEq(t, subset, 1, 64)
		javaSubsetNumEq(t, subset, 2, 2016)
		javaSubsetNumEq(t, subset, 62, 2016)
		javaSubsetNumEq(t, subset, 63, 64)
		javaSubsetNumEq(t, subset, 64, 1)
		javaSubsetNumEq(t, subset, 32, 1832624140942590534)
		_, err := NewSubsetValue(NewIntervalValue(1, 67)).NumberOfKElements(33)
		if _, ok := err.(*IllegalArgumentException); !ok {
			t.Fatalf("expected IllegalArgumentException for count exceeding Long.MAX_VALUE, got %T: %v", err, err)
		}
	})
	t.Run("testRandomSubsetGeneratorKNegative", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		javaFcnSetNull(t, javaSubsetRandomOrder(t, NewKSubsetValue(-1, NewIntervalValue(1, 2))))
	})
	t.Run("testRandomSubsetE5F01", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d", "e"), true)
		size := int(javaIntOneLeftShift(javaSubsetEnumerationSize(t, inner)))
		javaSubsetDoTest(t, size, inner, 4)
	})
	t.Run("testRandomSubsetE5F05", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d", "e"), true)
		size := int(javaIntOneLeftShift(javaSubsetEnumerationSize(t, inner)))
		javaSubsetDoTest(t, size, inner, 16)
	})
	t.Run("testRandomSubsetE7F05", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d", "e", "f", "g"), true)
		size := int(javaIntOneLeftShift(javaSubsetEnumerationSize(t, inner)))
		javaSubsetDoTest(t, size, inner, 64)
	})
	t.Run("testRandomSubsetE5F1", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d", "e"), true)
		size := int(javaIntOneLeftShift(javaSubsetEnumerationSize(t, inner)))
		javaSubsetDoTest(t, size, inner, size)
	})
	t.Run("testRandomSubsetE6F1", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d", "e", "f"), true)
		size := int(javaIntOneLeftShift(javaSubsetEnumerationSize(t, inner)))
		javaSubsetDoTest(t, size, inner, size)
	})
	t.Run("testRandomSubsetE7F1", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d", "e", "f", "g"), true)
		size := int(javaIntOneLeftShift(javaSubsetEnumerationSize(t, inner)))
		javaSubsetDoTest(t, size, inner, size)
	})
	t.Run("testKElementsMatchElementsNormalized", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		subset := NewSubsetValue(NewIntervalValue(1, 6))
		a, b := javaFcnSetAll(t, subset.ElementsNormalized()), javaSubsetAllK(t, subset)
		if len(a) != len(b) {
			t.Fatalf("list lengths differ: %d/%d", len(a), len(b))
		}
		for i, v := range a {
			javaFcnSetEquals(t, v, b[i])
		}
	})
	t.Run("testRandomSubsetSubset65536", func(t *testing.T) {
		SetRandomEnumerableGenerator(classRandom)
		t.Cleanup(func() { ResetRandomEnumerableValues() })
		inner := NewSetEnumValue(javaSubsetFixture(shuffle, "a", "b", "c", "d"), true)
		innerSubset := NewSubsetValue(inner)
		subset := NewSubsetValue(innerSubset)
		expected := int(javaIntOneLeftShift(javaSubsetEnumerationSize(t, innerSubset)))
		javaFcnSetSize(t, subset, expected)
		unique := &javaFcnSetHash{}
		e := javaSubsetRandom(t, subset, expected)
		for next := javaFcnSetNext(t, e); next != nil; next = javaFcnSetNext(t, e) {
			unique.add(t, next)
		}
		javaFcnSetHashSize(t, unique, expected)
	})
}
