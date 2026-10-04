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
 *   Ian Morris Nieves - added support for fingerprint stack trace
 ******************************************************************************/
// Complete original parameterized SubsetEnumeratorTest: twelve enumerable
// rows, both original methods, and all eleven fractions in each method.
package tlc

import (
	"fmt"
	"math"
	"testing"
)

func javaSubsetEnumerationSize(t *testing.T, v Value) int {
	t.Helper()
	n, err := v.Size()
	if err != nil {
		t.Fatal(err)
	}
	return n
}
func TestJavaSubsetEnumerator(t *testing.T) {
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
	// @Parameters builds every shared row before any method runs.
	params := []Value{NewIntervalValue(1, 10)}
	vec := NewValueVec(10)
	for _, ch := range "ABCDEFGHIJ" {
		vec.Add(MakeModelValue(fmt.Sprintf("%d", ch)))
	}
	params = append(params, NewSetEnumValueVec(vec, false))
	params = append(params, NewSetOfTuplesValue([]Value{NewIntervalValue(1, 5), NewIntervalValue(1, 5)}))
	params = append(params, NewSetOfTuplesValue([]Value{javaFcnSetEmpty(), javaFcnSetEmpty()}))
	params = append(params, NewUnionValue(NewSetEnumValue([]Value{NewIntervalValue(1, 5), NewIntervalValue(5, 11)}, true)))
	params = append(params, NewUnionValue(javaFcnSetEmpty()))
	params = append(params, NewSetOfFcnsValue(NewIntervalValue(2, 5), NewSetEnumValue(stringValues("a", "b", "c"), true)))
	params = append(params, NewSetOfFcnsValue(NewIntervalValue(3, 5), javaFcnSetEmpty()))
	params = append(params, NewSetOfFcnsValue(NewSetEnumValue([]Value{MakeModelValue("m1"), MakeModelValue("m2"), MakeModelValue("m3")}, true), NewSubsetValue(NewSetEnumValue(stringValues("a", "b", "c"), true))))
	domain := NewSetEnumValue(stringValues("A1", "A2", "A3"), true)
	rangeValue := NewSetEnumValue(stringValues("v1", "v2", "v3"), true)
	params = append(params, NewSetOfFcnsValue(domain, rangeValue))
	params = append(params, NewSubsetValue(NewSetEnumValue(stringValues("a", "b", "c"), true)))
	params = append(params, NewSubsetValue(javaFcnSetEmpty()))
	FP64Init() // Original @Parameters call after constructing all values.
	fractions := []float64{0, .1, .2, .3, .4, .55, .625, .775, .8, .9, 1}
	for i, enumerable := range params {
		t.Run(fmt.Sprintf("row%d", i), func(t *testing.T) {
			t.Run("testElementsInt", func(t *testing.T) {
				for _, fraction := range fractions {
					t.Run(fmt.Sprintf("fraction%g", fraction), func(t *testing.T) {
						k := int(math.Ceil(float64(javaSubsetEnumerationSize(t, enumerable)) * fraction))
						e, err := randomValueEnumeration(enumerable, k)
						if err != nil {
							t.Fatal(err)
						}
						values := javaFcnSetAll(t, e)
						if len(values) != k {
							t.Fatalf("values size=%d, want %d", len(values), k)
						}
						unique := &javaFcnSetHash{}
						for _, v := range values {
							unique.add(t, v)
						}
						javaFcnSetHashSize(t, unique, len(values))
						for _, v := range values {
							member, err := enumerable.Member(v)
							if err != nil || !member {
								t.Fatalf("member=%v/%v", member, err)
							}
						}
					})
				}
			})
			t.Run("testGetRandomSubset", func(t *testing.T) {
				for _, fraction := range fractions {
					t.Run(fmt.Sprintf("fraction%g", fraction), func(t *testing.T) {
						k := int(math.Ceil(float64(javaSubsetEnumerationSize(t, enumerable)) * fraction))
						enumValue, err := randomSubsetOfEnumerable(k, enumerable)
						if err != nil {
							t.Fatal(err)
						}
						javaFcnSetSize(t, enumValue, k)
						// Java constructs the HashSet with enumValue.size() as capacity.
						_ = javaSubsetEnumerationSize(t, enumValue)
						values := &javaFcnSetHash{}
						e := javaFcnSetElements(t, enumValue)
						for v := javaFcnSetNext(t, e); v != nil; v = javaFcnSetNext(t, e) {
							member, err := enumerable.Member(v)
							if err != nil || !member {
								t.Fatalf("member=%v/%v", member, err)
							}
							values.add(t, v)
						}
						want := javaSubsetEnumerationSize(t, enumValue)
						// Original explicitly constructs a second HashSet from the first.
						unique := &javaFcnSetHash{}
						for _, v := range values.values {
							unique.add(t, v)
						}
						javaFcnSetHashSize(t, unique, want)
					})
				}
			})
		})
	}
}
