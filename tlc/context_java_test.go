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

// Complete mechanical translation of tlc2.util.ContextTest.
// The concrete Go SymbolNode needs none of DummySymbolNode's unused abstract XML methods.
package tlc

import "testing"

func requireJavaContextLookup(t *testing.T, actual, expected any) {
	t.Helper()
	if actual != expected {
		t.Fatalf("Context.lookup = %v, want %v", actual, expected)
	}
}

func TestJavaContextLookupEmpty(t *testing.T) {

	requireJavaContextLookup(t, EmptyContext.Lookup(NewSymbolNode("Dummy")), nil)
	requireJavaContextLookup(t, EmptyContext.LookupCutoff(NewSymbolNode("Dummy"), true), nil)
	requireJavaContextLookup(t, EmptyContext.LookupCutoff(NewSymbolNode("Dummy"), false), nil)
	requireJavaContextLookup(t, EmptyContext.Lookup(nil), nil)
	requireJavaContextLookup(t, EmptyContext.LookupCutoff(nil, true), nil)
	requireJavaContextLookup(t, EmptyContext.LookupCutoff(nil, false), nil)
}

func TestJavaContextLookupBranch(t *testing.T) {

	ctx := BranchContext(EmptyContext)
	requireJavaContextLookup(t, ctx.Lookup(NewSymbolNode("Dummy")), nil)
	requireJavaContextLookup(t, ctx.LookupCutoff(NewSymbolNode("Dummy"), true), nil)
	requireJavaContextLookup(t, ctx.LookupCutoff(NewSymbolNode("Dummy"), false), nil)
	requireJavaContextLookup(t, ctx.Lookup(nil), nil)
	requireJavaContextLookup(t, ctx.LookupCutoff(nil, false), nil)
	requireJavaContextLookup(t, ctx.LookupCutoff(nil, true), nil)
}

func TestJavaContextLookupSymbolNodeNull(t *testing.T) {

	ctx := BranchContext(EmptyContext)
	requireJavaContextLookup(t, ctx.Lookup(nil), nil)
}

func TestJavaContextLookup(t *testing.T) {

	name := NewSymbolNode("ctx1")
	value := "value1"

	ctx1 := EmptyContext.Cons(name, value)
	branch := BranchContext(ctx1)
	ctx2 := branch.Cons(NewSymbolNode("ctx2"), "value2")
	ctx3 := ctx2.Cons(NewSymbolNode("ctx3"), "value3")

	requireJavaContextLookup(t, ctx3.Lookup(name), value)
}

func TestJavaContextLookupCutOffFalse(t *testing.T) {

	name := NewSymbolNode("ctx1")
	value := "value1"

	ctx1 := EmptyContext.Cons(name, value)
	branch := BranchContext(ctx1)
	ctx2 := branch.Cons(NewSymbolNode("ctx2"), "value2")
	ctx3 := ctx2.Cons(NewSymbolNode("ctx3"), "value3")

	requireJavaContextLookup(t, ctx3.LookupCutoff(name, false), value)
}

func TestJavaContextLookupCutOffTrue(t *testing.T) {

	name := NewSymbolNode("ctx1")
	value := "value1"

	ctx1 := EmptyContext.Cons(name, value)
	branch := BranchContext(ctx1)
	ctx2 := branch.Cons(NewSymbolNode("ctx2"), "value2")
	ctx3 := ctx2.Cons(NewSymbolNode("ctx3"), "value3")

	requireJavaContextLookup(t, ctx3.LookupCutoff(name, true), nil)
}

func TestJavaContextLookupWithAtBranching(t *testing.T) {

	name := NewSymbolNode("ctx1")
	value := "value1"

	ctx1 := EmptyContext.Cons(name, value)
	branch := BranchContext(ctx1)

	requireJavaContextLookup(t, branch.Lookup(name), value)
}

func TestJavaContextLookupWithCutOffFalseAtBranching(t *testing.T) {

	name := NewSymbolNode("ctx1")
	value := "value1"

	ctx1 := EmptyContext.Cons(name, value)
	branch := BranchContext(ctx1)

	requireJavaContextLookup(t, branch.LookupCutoff(name, false), value)
}

func TestJavaContextLookupWithCutOffTrueAtBranching(t *testing.T) {

	name := NewSymbolNode("ctx1")
	value := "value1"

	ctx1 := EmptyContext.Cons(name, value)
	branch := BranchContext(ctx1)

	requireJavaContextLookup(t, branch.LookupCutoff(name, true), nil)
}

func TestJavaContextLookupSymbolNode(t *testing.T) {

	name := NewSymbolNode("Dummy")
	value := new(int)

	ctx := BranchContext(EmptyContext)
	cons := ctx.Cons(name, value)

	lookup := cons.Lookup(name)
	requireJavaContextLookup(t, lookup, value)
}
