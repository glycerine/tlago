/*******************************************************************************
 * Copyright (c) 20178 Microsoft Research. All rights reserved.
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
package tlc

import (
	"strings"
	"testing"
)

func javaRandomizationEvalMessage(t *testing.T, want string, body func() (Value, error)) {
	t.Helper()
	_, err := body()
	if err == nil {
		t.Fatal("fail(): no EvalException")
	}
	if !isValueEvalException(err) {
		panic(err)
	}
	message := err.(interface{ GetMessage() *string }).GetMessage()
	if !strings.Contains(*message, want) {
		t.Fatalf("assertTrue(getMessage().contains(%q)): %q", want, *message)
	}
}

func javaRandomizationEnumerable(value Value, err error) Value {
	if err != nil {
		panic(err)
	}
	// Source casts to Enumerable rather than asserting an extra instanceof.
	if value != nil {
		if _, ok := asEnumerable(value); !ok {
			panic(NewClassCastException("Enumerable"))
		}
	}
	return value
}

func javaRandomizationSize(t *testing.T, want int, value Value) {
	t.Helper()
	got, err := value.Size()
	if err != nil {
		panic(err)
	}
	if got != want {
		t.Fatalf("size %d, want %d", got, want)
	}
}

func javaRandomizationNotNull(t *testing.T, value Value) {
	t.Helper()
	if value == nil {
		t.Fatal("assertNotNull(randomSubset)")
	}
}

func javaRandomizationMember(t *testing.T, set, member Value) {
	t.Helper()
	got, err := set.Member(member)
	if err != nil {
		panic(err)
	}
	if !got {
		t.Fatal("assertTrue(randomSubset.member(...))")
	}
}

// Entire original RandomizationTest, including duplicate zero/negative cases.
func TestJavaRandomization(t *testing.T) {
	oldPoly, oldSeed := FP64IrredPoly(), RandomEnumerableSeed()
	oldRng := ResetRandomEnumerableValues()
	oldChecker, oldSimulator := MainChecker(), CurrentSimulator()
	SetMainChecker(nil)
	SetSimulator(nil)
	t.Cleanup(func() {
		FP64InitPoly(oldPoly)
		SetRandomEnumerableSeed(oldSeed)
		SetRandomEnumerableGenerator(oldRng)
		SetMainChecker(oldChecker)
		SetSimulator(oldSimulator)
	})
	// Original @BeforeClass. JUnit runs every method on the same thread.
	SetRandomEnumerableSeed(15041980)
	FP64Init()
	classRandom := RandomEnumerableGenerator()
	run := func(name string, body func(*testing.T)) {
		t.Run(name, func(t *testing.T) {
			SetRandomEnumerableGenerator(classRandom)
			t.Cleanup(func() { ResetRandomEnumerableValues() })
			body(t)
		})
	}
	// JUnit MethodSorters.DEFAULT: signed String.hashCode, then method name.
	run("testSetNonFinite", func(t *testing.T) {
		javaRandomizationEvalMessage(t, "The third argument of RandomSubsetSetProbability should be a finite set, but instead it is:\nNat", func() (Value, error) {
			return RandomSubsetSet(NewIntValue(23), NewStringValue("1.0"), Nat())
		})
	})
	run("testRandomSubsetNonFinite", func(t *testing.T) {
		javaRandomizationEvalMessage(t, "The second argument of RandomSubset should be a a finite set, but instead it is:\nNat", func() (Value, error) {
			return RandomSubset(NewIntValue(23), Nat())
		})
	})
	run("testRSSV2Zero", func(t *testing.T) {
		randomSubset := javaRandomizationEnumerable(RandomSetOfSubsets(NewIntValue(23), NewIntValue(0), NewIntervalValue(1, 42)))
		javaRandomizationSize(t, 1, randomSubset)
		javaRandomizationMember(t, randomSubset, NewSetEnumValue(nil, true))
	})
	run("testV2Negative", func(t *testing.T) {
		javaRandomizationEvalMessage(t, "The second argument of RandomSetOfSubsets should be a nonnegative integer, but instead it is:\n-1", func() (Value, error) {
			return RandomSetOfSubsets(NewIntValue(23), NewIntValue(-1), NewIntervalValue(1, 42))
		})
	})
	run("testV1Valid", func(t *testing.T) {
		randomSubset := javaRandomizationEnumerable(RandomSubsetSet(NewIntValue(42), NewStringValue("0.1"), NewIntervalValue(1, 42)))
		javaRandomizationNotNull(t, randomSubset)
		javaRandomizationSize(t, 42, randomSubset)
	})
	run("testV3Empty", func(t *testing.T) {
		javaRandomizationEvalMessage(t, "The first argument of RandomSetOfSubsets should be a nonnegative integer that is smaller than the subset's size of 2^0, but instead it is:\n42", func() (Value, error) {
			return RandomSetOfSubsets(NewIntValue(42), NewIntValue(42), NewSetEnumValue(nil, true))
		})
	})
	run("testRSSV2Negative", func(t *testing.T) {
		javaRandomizationEvalMessage(t, "The second argument of RandomSetOfSubsets should be a nonnegative integer, but instead it is:\n-1", func() (Value, error) {
			return RandomSetOfSubsets(NewIntValue(23), NewIntValue(-1), NewIntervalValue(1, 42))
		})
	})
	run("testRSSV2Cardinality", func(t *testing.T) {
		randomSubset := javaRandomizationEnumerable(RandomSetOfSubsets(NewIntValue(32), NewIntValue(5), NewIntervalValue(1, 5)))
		javaRandomizationSize(t, 1, randomSubset)
		// Source probability 1: all generated subsets collide with the input.
		javaRandomizationMember(t, randomSubset, NewIntervalValue(1, 5))
	})
	run("testV3AstronomicallyLarge", func(t *testing.T) {
		randomSubset := javaRandomizationEnumerable(RandomSetOfSubsets(NewIntValue(42), NewIntValue(42), NewIntervalValue(1, 256)))
		javaRandomizationNotNull(t, randomSubset)
		javaRandomizationSize(t, 42, randomSubset)
	})
	run("testV1Negative", func(t *testing.T) {
		var v1 Value = NewIntValue(-42)
		javaRandomizationEvalMessage(t, "The first argument of RandomSetOfSubsets should be a nonnegative integer, but instead it is:\n-42", func() (Value, error) {
			return RandomSetOfSubsets(v1, NewIntValue(42), NewIntervalValue(1, 42))
		})
	})
	run("testV1NoIntValue", func(t *testing.T) {
		var v1 Value = NewStringValue("52")
		javaRandomizationEvalMessage(t, "The first argument of RandomSetOfSubsets should be a nonnegative integer, but instead it is:\n\"52\"", func() (Value, error) {
			return RandomSetOfSubsets(v1, NewIntValue(42), NewIntervalValue(1, 42))
		})
	})
	run("testV2Larger1", func(t *testing.T) {
		javaRandomizationEvalMessage(t, "1.1", func() (Value, error) {
			return RandomSubsetSet(NewIntValue(23), NewStringValue("1.1"), NewIntervalValue(1, 42))
		})
	})
	run("testV3isInfinite", func(t *testing.T) {
		javaRandomizationEvalMessage(t, "The third argument of RandomSetOfSubsets should be a finite set, but instead it is:\nNat", func() (Value, error) {
			return RandomSetOfSubsets(NewIntValue(42), NewIntValue(42), Nat())
		})
	})
	run("testV1Zero", func(t *testing.T) {
		var v1 Value = NewIntValue(0)
		randomSubset := javaRandomizationEnumerable(RandomSetOfSubsets(v1, NewIntValue(42), NewIntervalValue(1, 42)))
		javaRandomizationNotNull(t, randomSubset)
		javaRandomizationSize(t, 0, randomSubset)
	})
	run("testV2Zero", func(t *testing.T) {
		randomSubset := javaRandomizationEnumerable(RandomSetOfSubsets(NewIntValue(23), NewIntValue(0), NewIntervalValue(1, 42)))
		javaRandomizationSize(t, 1, randomSubset)
		javaRandomizationMember(t, randomSubset, NewSetEnumValue(nil, true))
	})
	run("testRSSV2TwiceCardinality", func(t *testing.T) {
		javaRandomizationEvalMessage(t, "The second argument of RandomSetOfSubsets should be a nonnegative integer in range 0..Cardinality(S), but instead it is:\n10", func() (Value, error) {
			return RandomSetOfSubsets(NewIntValue(23), NewIntValue(10), NewIntervalValue(1, 5))
		})
	})
}
