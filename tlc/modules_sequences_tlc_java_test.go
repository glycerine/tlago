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
// Portions Copyright (c) 2016 Microsoft Research. All rights reserved.
package tlc

import (
	"math"
	"testing"
)

func javaSequencesStringEquals(t *testing.T, want string, value Value, err error) {
	t.Helper()
	if err != nil {
		panic(err)
	}
	s, ok := value.(*StringValue)
	if !ok {
		t.Fatalf("assertTrue(value instanceof StringValue): %T", value)
	}
	if !UniqueStringOf(want).Equal(s.Val) {
		t.Fatalf("UniqueString expected %q, got %q", want, s.Val.String())
	}
}

// Native modules retain the established TLCError EvalException carrier;
// its Runtime flag distinguishes TLCRuntimeException and must be excluded.
func javaSequencesEvalCode(t *testing.T, want int, body func() (Value, error)) {
	t.Helper()
	_, err := body()
	if err == nil {
		t.Fatal("fail(): no EvalException")
	}
	if !isValueEvalException(err) {
		panic(err)
	}
	var code int
	switch e := err.(type) {
	case *EvalException:
		code = e.GetErrorCode()
	case *TLCError:
		code = e.Code
	default:
		panic(err)
	}
	if code != want {
		t.Fatalf("EvalException code %d, want %d", code, want)
	}
}

// Entire original SequencesTest, including each named error-code catch.
func TestJavaSequences(t *testing.T) {
	t.Run("testTailString", func(t *testing.T) { v, e := Tail(NewStringValue("abc")); javaSequencesStringEquals(t, "bc", v, e) })
	t.Run("testHeadString", func(t *testing.T) {
		javaSequencesEvalCode(t, ECTLCModuleOneArgumentError, func() (Value, error) { return Head(NewStringValue("a")) })
	})
	t.Run("testHeadStringEmpty", func(t *testing.T) {
		javaSequencesEvalCode(t, ECTLCModuleOneArgumentError, func() (Value, error) { return Head(NewStringValue("")) })
	})
	t.Run("testAppendString", func(t *testing.T) {
		javaSequencesEvalCode(t, ECTLCModuleEvaluating, func() (Value, error) { return Append(NewStringValue(""), NewStringValue("a")) })
	})
	t.Run("testAppendString2", func(t *testing.T) {
		javaSequencesEvalCode(t, ECTLCModuleEvaluating, func() (Value, error) { return Append(NewStringValue("abc"), NewStringValue("d")) })
	})
	t.Run("testAppendStringNonString", func(t *testing.T) {
		javaSequencesEvalCode(t, ECTLCModuleEvaluating, func() (Value, error) { return Append(NewStringValue(""), IntZero) })
	})
	t.Run("testConcatStringToSeq", func(t *testing.T) {
		javaSequencesEvalCode(t, ECTLCModuleEvaluating, func() (Value, error) {
			return Concat(NewTupleValue([]Value{NewStringValue("abc")}), NewStringValue("d"))
		})
	})
	t.Run("testConcatSeqToString", func(t *testing.T) {
		javaSequencesEvalCode(t, ECTLCModuleEvaluating, func() (Value, error) {
			return Concat(NewStringValue("abc"), NewTupleValue([]Value{NewStringValue("d")}))
		})
	})
	t.Run("testConcatStringToString", func(t *testing.T) {
		v, e := Concat(NewStringValue("abc"), NewStringValue("d"))
		javaSequencesStringEquals(t, "abcd", v, e)
	})
	t.Run("testConcatIntToSeq", func(t *testing.T) {
		javaSequencesEvalCode(t, ECTLCModuleEvaluating, func() (Value, error) { return Concat(NewTupleValue([]Value{NewStringValue("abc")}), IntOne) })
	})
	t.Run("testConcatSeqToInt", func(t *testing.T) {
		javaSequencesEvalCode(t, ECTLCModuleEvaluating, func() (Value, error) { return Concat(IntOne, NewTupleValue([]Value{NewStringValue("d")})) })
	})
	t.Run("testConcatIntToInt", func(t *testing.T) {
		javaSequencesEvalCode(t, ECTLCModuleEvaluating, func() (Value, error) { return Concat(IntOne, IntOne) })
	})
	t.Run("testSubseq", func(t *testing.T) {
		v, e := SubSeq(NewStringValue("abc"), IntOne, IntOne)
		javaSequencesStringEquals(t, "a", v, e)
	})
}

func javaTLCModuleIntEquals(t *testing.T, want, got int) {
	t.Helper()
	if want != got {
		t.Fatalf("got %d, want %d", got, want)
	}
}
func javaTLCModuleArrayEquals(t *testing.T, want, got []Value) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("array length %d, want %d", len(got), len(want))
	}
	for i, expected := range want {
		equal, err := expected.Equal(got[i])
		if err != nil {
			panic(err)
		}
		if !equal {
			t.Fatalf("array element %d: got %v, want %v", i, got[i], expected)
		}
	}
}

// Entire tlc2.module.TLCTest, with source @BeforeClass FP64 initialization.
func TestJavaTLCModule(t *testing.T) {
	FP64Init()
	t.Run("testA", func(t *testing.T) {
		f := NewFcnRcdValue([]Value{NewIntValue(3)}, []Value{NewIntValue(11)}, true)
		g := NewTupleValue([]Value{NewIntValue(1), NewIntValue(2), NewIntValue(3)})
		combined, e := CombineFcn(f, g)
		if e != nil {
			panic(e)
		}
		r, ok := combined.(*FcnRcdValue)
		if !ok {
			t.Fatal("assertTrue(combined instanceof FcnRcdValue)")
		}
		r.Normalize()
		javaTLCModuleIntEquals(t, 3, len(r.Domain))
		javaTLCModuleArrayEquals(t, []Value{NewIntValue(1), NewIntValue(2), NewIntValue(3)}, r.Domain)
		javaTLCModuleIntEquals(t, 3, len(r.Values))
		javaTLCModuleArrayEquals(t, []Value{NewIntValue(1), NewIntValue(2), NewIntValue(11)}, r.Values)
	})
	t.Run("testB", func(t *testing.T) {
		f := NewFcnRcdValue([]Value{NewIntValue(4)}, []Value{NewIntValue(11)}, true)
		g := NewTupleValue([]Value{NewIntValue(1), NewIntValue(2), NewIntValue(3), NewIntValue(11)})
		combined, e := CombineFcn(f, g)
		if e != nil {
			panic(e)
		}
		r, ok := combined.(*FcnRcdValue)
		if !ok {
			t.Fatal("assertTrue(combined instanceof FcnRcdValue)")
		}
		r.Normalize()
		javaTLCModuleIntEquals(t, 4, len(r.Domain))
		javaTLCModuleArrayEquals(t, []Value{NewIntValue(1), NewIntValue(2), NewIntValue(3), NewIntValue(4)}, r.Domain)
		javaTLCModuleIntEquals(t, 4, len(r.Values))
		javaTLCModuleArrayEquals(t, []Value{NewIntValue(1), NewIntValue(2), NewIntValue(3), NewIntValue(11)}, r.Values)
	})
	maxIntCombine := func(t *testing.T, f, g Value) {
		combined, e := CombineFcn(f, g)
		if e != nil {
			panic(e)
		}
		// Source casts here rather than asserting an extra instanceof.
		r := combined.(*FcnRcdValue)
		javaTLCModuleArrayEquals(t, []Value{NewIntValue(math.MaxInt32)}, r.Domain)
		javaTLCModuleArrayEquals(t, []Value{IntZero}, r.Values)
	}
	t.Run("testCombineMaxIntIntervalOnLeft", func(t *testing.T) {
		f := NewFcnRcdIntervalValue(NewIntervalValue(math.MaxInt32, math.MaxInt32), []Value{IntZero})
		g := NewFcnRcdValue([]Value{NewIntValue(math.MaxInt32)}, []Value{IntOne}, true)
		maxIntCombine(t, f, g)
	})
	t.Run("testCombineMaxIntIntervalOnRight", func(t *testing.T) {
		f := NewFcnRcdValue([]Value{NewIntValue(math.MaxInt32)}, []Value{IntZero}, true)
		g := NewFcnRcdIntervalValue(NewIntervalValue(math.MaxInt32, math.MaxInt32), []Value{IntOne})
		maxIntCombine(t, f, g)
	})
	t.Run("testPermutations", func(t *testing.T) {
		in := NewIntervalValue(1, 5).ToSetEnum()
		n, e := in.Size()
		if e != nil {
			panic(e)
		}
		if n != 5 {
			t.Fatalf("input size %d, want 5", n)
		}
		permutations, e := Permutations(in)
		if e != nil {
			panic(e)
		}
		enum, ok := asEnumerable(permutations)
		if !ok {
			t.Fatal("assertTrue(permutations instanceof Enumerable)")
		}
		size, e := permutations.Size()
		if e != nil {
			panic(e)
		}
		if size != 120 {
			t.Fatalf("permutations size %d, want 120", size)
		}
		// Reuse the original Value.hashCode/equals HashSet adapter, retaining actual
		// Value keys. A string-rendering deduplicator would weaken this assertion.
		seen := javaFcnSetHash{buckets: make(map[int32][]Value, size)}
		elements := enum.Elements()
		for {
			v := elements.NextElement()
			if e := elements.Err(); e != nil {
				panic(e)
			}
			if v == nil {
				break
			}
			want, e := in.Size()
			if e != nil {
				panic(e)
			}
			got, e := v.Size()
			if e != nil {
				panic(e)
			}
			if got != want {
				t.Fatalf("value size %d, want %d", got, want)
			}
			seen.add(t, v)
		}
		if len(seen.values) != 120 {
			t.Fatalf("HashSet size %d, want 120", len(seen.values))
		}
	})
}
