/*******************************************************************************
 * Copyright (c) 2015 Microsoft Research. All rights reserved.
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

// Translation of both complete RecordValueTest methods, retaining the original
// conditional catch assertions and isolated InternTable(2).
package tlc

import (
	"strings"
	"testing"
)

func TestRecordValueDeepCopyIndependentOfSourceNormalization(t *testing.T) {
	assertTrue := func(condition bool) {
		t.Helper()
		if !condition {
			t.Fatal("Java RecordValue assertion failed")
		}
	}
	internTable := NewInternTable(2)
	a := internTable.Put("a")
	b := internTable.Put("b")
	aVal := NewStringValue("aVal")
	bVal := NewStringValue("bVal")
	orig := NewRecordValue([]*UniqueString{b, a}, []Value{bVal, aVal}, false)
	assertTrue(orig.Names[0].Equal(b))
	assertTrue(orig.Names[1].Equal(a))
	assertTrue(mustValueEqual(orig.Values[0], bVal))
	assertTrue(mustValueEqual(orig.Values[1], aVal))
	deepCopy := orig.DeepCopy().(*RecordValue)
	orig.DeepNormalize()
	assertTrue(orig.Names[0].Equal(a))
	assertTrue(orig.Names[1].Equal(b))
	assertTrue(mustValueEqual(orig.Values[0], aVal))
	assertTrue(mustValueEqual(orig.Values[1], bVal))
	assertTrue(deepCopy.Names[0].Equal(b))
	assertTrue(deepCopy.Names[1].Equal(a))
	assertTrue(mustValueEqual(deepCopy.Values[0], bVal))
	assertTrue(mustValueEqual(deepCopy.Values[1], aVal))
}

func TestRecordValueApplyReportsMissingAndNonStringFields(t *testing.T) {
	assertTrue := func(condition bool) {
		t.Helper()
		if !condition {
			t.Fatal("Java RecordValue assertion failed")
		}
	}
	aVal := NewStringValue("aVal")
	recVal := NewRecordValue([]*UniqueString{UniqueStringOf("a")}, []Value{aVal}, false)
	if _, err := recVal.Apply(NewStringValue("b")); err != nil {
		if !isJavaEvalOrRuntimeException(err) {
			t.Fatal(err)
		}
		assertTrue(strings.Contains(err.Error(), "Attempted to access nonexistent field 'b' of record\n[a |-> \"aVal\"]"))
	}
	if _, err := recVal.Apply(NewIntValue(0)); err != nil {
		if !isJavaEvalOrRuntimeException(err) {
			t.Fatal(err)
		}
		assertTrue(strings.Contains(err.Error(), "Attempted to access record by a non-string argument: 0"))
	}
	if _, err := recVal.Select(NewIntValue(0)); err != nil {
		if !isJavaEvalOrRuntimeException(err) {
			t.Fatal(err)
		}
		assertTrue(strings.Contains(err.Error(), "Attempted to access record by a non-string argument: 0"))
	}
}
