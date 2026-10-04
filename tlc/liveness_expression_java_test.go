/*******************************************************************************
 * Copyright (c) 2021 Microsoft Research. All rights reserved.
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

import "testing"

// TBParTest.DummyOpApplNode preserves its named predicate and String override.
type javaLivenessDummyOpAppl struct{ *OpApplNode }

func (n *javaLivenessDummyOpAppl) String() string { return n.Operator.Name.String() }
func javaLivenessDummyState(name string) *LiveExprNode {
	p := &javaLivenessDummyOpAppl{NewOpApplNode(NewSymbolNode(name))}
	return NewLNState(p.String(), p, EmptyContext, nil)
}

// Original LiveExprNodeTest.testLNBool; retain source LNNext TODO assertions as comments.
func TestJavaLiveExprNodeLNBool(t *testing.T) {

	// Rewriting rules page 452.
	p := NewLNBool(true)
	q := NewLNBool(true)

	// ~~T -> T
	tf := NewLNNeg(NewLNNeg(p))
	if tf.IsPositiveForm() {
		t.Fatal("original assertion 1 failed")
	}
	if !tf.PushNeg().IsPositiveForm() {
		t.Fatal("original assertion 2 failed")
	}
	if !p.Equal(tf.ToDNF()) {
		t.Fatal("original assertion 3 failed")
	}

	// ~<>T -> []F
	tf = NewLNNeg(NewLNEven(p))
	if tf.IsPositiveForm() {
		t.Fatal("original assertion 4 failed")
	}
	if !tf.PushNeg().IsPositiveForm() {
		t.Fatal("original assertion 5 failed")
	}
	if !NewLNAll(NewLNBool(false)).Equal(tf.ToDNF()) {
		t.Fatal("original assertion 6 failed")
	}

	// ~[]T -> <>F
	tf = NewLNNeg(NewLNAll(p))
	if tf.IsPositiveForm() {
		t.Fatal("original assertion 7 failed")
	}
	if !tf.PushNeg().IsPositiveForm() {
		t.Fatal("original assertion 8 failed")
	}
	if !NewLNEven(NewLNBool(false)).Equal(tf.ToDNF()) {
		t.Fatal("original assertion 9 failed")
	}

	// ~()T -> ()F
	//TODO: pushNeg causes a StackOverflow, see LiveExprNode#pushNeg
	tf = NewLNNeg(NewLNNext(p))
	if tf.IsPositiveForm() {
		t.Fatal("original assertion 10 failed")
	}
	//		assertTrue(tf.PushNeg().IsPositiveForm())
	//		assertTrue(NewLNNext(NewLNBool(false)).Equal(tf.ToDNF()))

	// ~(T /\ T) -> F \/ F
	//TODO: Rewrite F\/F to F in toDNF? Probably not worth it.
	tf = NewLNNeg(NewLNConj(p, q))
	if tf.IsPositiveForm() {
		t.Fatal("original assertion 11 failed")
	}
	if !tf.PushNeg().IsPositiveForm() {
		t.Fatal("original assertion 12 failed")
	}
	if !NewLNDisj(NewLNBool(false), NewLNBool(false)).Equal(tf.ToDNF()) {
		t.Fatal("original assertion 13 failed")
	}

	// ~(T \/ T) -> F /\ F
	//TODO: Rewrite F/\F to F in toDNF? Probably not worth it.
	tf = NewLNNeg(NewLNDisj(p, q))
	if tf.IsPositiveForm() {
		t.Fatal("original assertion 14 failed")
	}
	if !tf.PushNeg().IsPositiveForm() {
		t.Fatal("original assertion 15 failed")
	}
	if !NewLNConj(NewLNBool(false), NewLNBool(false)).Equal(tf.ToDNF()) {
		t.Fatal("original assertion 16 failed")
	}

}

// Original LiveExprNodeTest.testLNState; retain source LNNext TODO assertions as comments.
func TestJavaLiveExprNodeLNState(t *testing.T) {

	p := javaLivenessDummyState("p")
	q := javaLivenessDummyState("q")

	// ~~p -> p
	tf := NewLNNeg(NewLNNeg(p))
	if tf.IsPositiveForm() {
		t.Fatal("original assertion 1 failed")
	}
	if !tf.PushNeg().IsPositiveForm() {
		t.Fatal("original assertion 2 failed")
	}
	if !p.Equal(tf.ToDNF()) {
		t.Fatal("original assertion 3 failed")
	}

	// ~<>p -> []~p
	tf = NewLNNeg(NewLNEven(p))
	if tf.IsPositiveForm() {
		t.Fatal("original assertion 4 failed")
	}
	if !tf.PushNeg().IsPositiveForm() {
		t.Fatal("original assertion 5 failed")
	}
	if !NewLNAll(NewLNNeg(p)).Equal(tf.ToDNF()) {
		t.Fatal("original assertion 6 failed")
	}

	// ~[]p -> <>~p
	tf = NewLNNeg(NewLNAll(p))
	if tf.IsPositiveForm() {
		t.Fatal("original assertion 7 failed")
	}
	if !tf.PushNeg().IsPositiveForm() {
		t.Fatal("original assertion 8 failed")
	}
	if !NewLNEven(NewLNNeg(p)).Equal(tf.ToDNF()) {
		t.Fatal("original assertion 9 failed")
	}

	// ~()p -> ()~p
	//TODO: pushNeg causes a StackOverflow, see LiveExprNode#pushNeg
	tf = NewLNNeg(NewLNNext(p))
	if tf.IsPositiveForm() {
		t.Fatal("original assertion 10 failed")
	}
	//		assertTrue(tf.PushNeg().IsPositiveForm())
	//		assertTrue(NewLNNext(NewLNNeg(p)).Equal(tf.ToDNF()))

	// ~(p /\ q) -> ~p \/ ~q
	tf = NewLNNeg(NewLNConj(p, q))
	if tf.IsPositiveForm() {
		t.Fatal("original assertion 11 failed")
	}
	if !tf.PushNeg().IsPositiveForm() {
		t.Fatal("original assertion 12 failed")
	}
	if !NewLNDisj(NewLNNeg(p), NewLNNeg(q)).Equal(tf.ToDNF()) {
		t.Fatal("original assertion 13 failed")
	}

	// ~(p \/ q) -> ~p /\ ~q
	tf = NewLNNeg(NewLNDisj(p, q))
	if tf.IsPositiveForm() {
		t.Fatal("original assertion 14 failed")
	}
	if !tf.PushNeg().IsPositiveForm() {
		t.Fatal("original assertion 15 failed")
	}
	if !NewLNConj(NewLNNeg(p), NewLNNeg(q)).Equal(tf.ToDNF()) {
		t.Fatal("original assertion 16 failed")
	}

}
