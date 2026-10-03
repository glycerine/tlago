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
// Ports of SetOfTuplesValueTest's three indexed sampling methods.
// Its lazy rendering and empty ordinary enumeration methods are already ported
// in value_setconstructors_test.go.
package tlc

import "testing"

func TestJavaSetOfTuplesValueIndexedSampling(t *testing.T) {
	oldSeed := RandomEnumerableSeed()
	oldRandom := ResetRandomEnumerableValues()
	oldPoly := FP64IrredPoly()
	oldSetBound := Globals.SetBound
	SetRandomEnumerableSeed(15041980)
	FP64Init()
	Globals.SetBound = 1000000
	t.Cleanup(func() {
		SetRandomEnumerableSeed(oldSeed)
		SetRandomEnumerableGenerator(oldRandom)
		FP64InitPoly(oldPoly)
		Globals.SetBound = oldSetBound
	})

	t.Run("testRandomSubsetEmptyNonEnumerableComponent", func(t *testing.T) {
		product := NewSetOfTuplesValue([]Value{Nat(), EmptySet})
		if size, err := product.Size(); err != nil || size != 0 {
			t.Fatalf("product size=%d/%v, want 0", size, err)
		}
		if big, err := randomProductNeedsBigInteger(product); err != nil || big {
			t.Fatalf("needBigInteger=%v/%v, want false", big, err)
		}
		elements, err := randomValueEnumeration(product, 1)
		if err != nil {
			t.Fatal(err)
		}
		if size := countValues(t, elements); size != 0 {
			t.Fatalf("sample size=%d, want 0", size)
		}
		subset, _, err := randomSubsetOfProductValue(1, product)
		if err != nil {
			t.Fatal(err)
		}
		if size, err := subset.Size(); err != nil || size != 0 {
			t.Fatalf("random subset size=%d/%v, want 0", size, err)
		}
	})

	t.Run("testRandomSubsetBeyondSetBound", func(t *testing.T) {
		interval := NewIntervalValue(1, 200)
		product := NewSetOfTuplesValue([]Value{interval, interval, interval})
		if big, err := randomProductNeedsBigInteger(product); err != nil || big {
			t.Fatalf("needBigInteger=%v/%v, want false", big, err)
		}
		if size, err := product.Size(); err != nil || size != 8000000 || size <= Globals.SetBound {
			t.Fatalf("product size=%d/%v, want 8000000 > setBound", size, err)
		}
		elements, err := randomValueEnumeration(product, 1000)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := elements.(*randomIndexedValueEnumeration); !ok {
			t.Fatalf("enumerator=%T, want source SubsetEnumerator", elements)
		}
		requireJavaTupleProductRandomSubset(t, product, 1000)
	})

	t.Run("testRandomSubsetBeyondIntMaxValue", func(t *testing.T) {
		interval := NewIntervalValue(1, 4000)
		product := NewSetOfTuplesValue([]Value{interval, interval, interval})
		if big, err := randomProductNeedsBigInteger(product); err != nil || !big {
			t.Fatalf("needBigInteger=%v/%v, want true", big, err)
		}
		if _, err := product.Size(); javaRuntimeException(err) == nil {
			t.Fatalf("size error=%v, want source TLCRuntimeException", err)
		}
		elements, err := randomValueEnumeration(product, 1000)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := elements.(*randomBigProductEnumeration); !ok {
			t.Fatalf("enumerator=%T, want source BigIntegerSubsetEnumerator", elements)
		}
		requireJavaTupleProductRandomSubset(t, product, 1000)
	})
}

// Source assertRandomSubset checks list size, HashSet size, every membership,
// and the size returned by getRandomSubset. Hash buckets retain Value.equals.
func requireJavaTupleProductRandomSubset(t *testing.T, product *SetOfTuplesValue, k int) {
	t.Helper()
	elements, err := randomValueEnumeration(product, k)
	if err != nil {
		t.Fatal(err)
	}
	values := []Value{}
	for value := elements.NextElement(); value != nil; value = elements.NextElement() {
		values = append(values, value)
	}
	if err := elements.Err(); err != nil {
		t.Fatal(err)
	}
	if len(values) != k {
		t.Fatalf("sample size=%d, want %d", len(values), k)
	}
	distinct := 0
	buckets := map[int32][]Value{}
	for _, value := range values {
		hash := ValueJavaHashCode(value)
		duplicate := false
		for _, saved := range buckets[hash] {
			equal, err := value.Equal(saved)
			if err != nil {
				t.Fatal(err)
			}
			if equal {
				duplicate = true
				break
			}
		}
		if !duplicate {
			buckets[hash] = append(buckets[hash], value)
			distinct++
		}
	}
	if distinct != k {
		t.Fatalf("distinct sample size=%d, want %d", distinct, k)
	}
	for _, value := range values {
		if member, err := product.Member(value); err != nil || !member {
			t.Fatalf("product member=%v/%v for %v", member, err, value)
		}
	}
	subset, _, err := randomSubsetOfProductValue(k, product)
	if err != nil {
		t.Fatal(err)
	}
	if size, err := subset.Size(); err != nil || size != k {
		t.Fatalf("random subset size=%d/%v, want %d", size, err, k)
	}
}
