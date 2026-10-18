/*******************************************************************************
 * Copyright (c) 2020 Microsoft Research. All rights reserved.
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
 ******************************************************************************/
// Complete original TupleValueTest.testErrorMessages.
package tlc

import (
	"strings"
	"testing"
)

// The original catches TLCRuntimeException and checks contains only when thrown;
// it has no fail() after these try blocks. Other exceptions remain uncaught.
func javaTupleCaughtMessage(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		return
	}
	failure := javaRuntimeException(err)
	if failure == nil {
		t.Fatalf("uncaught %T: %v", err, err)
	}
	if !strings.Contains(failure.Error(), want) {
		t.Fatalf("message=%q, want contains %q", failure.Error(), want)
	}
}

func TestJavaTupleValue(t *testing.T) {
	t.Run("testErrorMessages", func(t *testing.T) {
		elems := []Value{NewStringValue("A")}
		tupVal := NewTupleValue(elems)
		_, err := tupVal.Apply(NewIntValue(2))
		javaTupleCaughtMessage(t, err, "Attempted to access index 2 of tuple\n<<\"A\">>\nwhich is out of bounds")
		_, err = tupVal.Apply(NewStringValue("a"))
		javaTupleCaughtMessage(t, err, "Attempted to access tuple at a non integral index: \"a\"")
		_, err = tupVal.Select(NewStringValue("a"))
		javaTupleCaughtMessage(t, err, "Attempted to access tuple at a non integral index: \"a\"")
		args := []Value{NewStringValue("arg1"), NewStringValue("arg2")}
		_, err = tupVal.ApplyArgs(args, EvalClear)
		javaTupleCaughtMessage(t, err, "Attempted to access tuple with 2 arguments when it expects 1.")
	})
}
