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

// Original TBParTest.testParticleClosureInconsistentConstantLevel.
func TestJavaTBParParticleClosureInconsistentConstantLevel(t *testing.T) {
	tbPar := NewTBPar(0)
	tbPar.AddElement(NewLNBool(false))
	tbPar.AddElement(NewLNNeg(NewLNBool(false)))
	if got := tbPar.ParticleClosure().Size(); got != 0 {
		t.Fatalf("particleClosure size=%d, want0", got)
	}
}

// Original TBParTest.testParticleClosureInconsistentStateLevel.
func TestJavaTBParParticleClosureInconsistentStateLevel(t *testing.T) {
	tbPar := NewTBPar(0)
	p := javaLivenessDummyState("p")
	tbPar.AddElement(p)
	tbPar.AddElement(NewLNNeg(p))
	if got := tbPar.ParticleClosure().Size(); got != 0 {
		t.Fatalf("particleClosure size=%d, want0", got)
	}
}

// Original TBParTest.testParticleClosureConsistentConstantLevel.
func TestJavaTBParParticleClosureConsistentConstantLevel(t *testing.T) {
	tbPar := NewTBPar(0)
	tbPar.AddElement(NewLNBool(false))
	tbPar.AddElement(NewLNNeg(NewLNBool(true)))
	if got := tbPar.ParticleClosure().Size(); got != 1 {
		t.Fatalf("particleClosure size=%d, want1", got)
	}
}

// Original TBParTest.testParticleClosureConsistentStateLevel.
func TestJavaTBParParticleClosureConsistentStateLevel(t *testing.T) {
	tbPar := NewTBPar(0)
	p := javaLivenessDummyState("p")
	tbPar.AddElement(p)
	tbPar.AddElement(NewLNNeg(NewLNNeg(p)))
	if got := tbPar.ParticleClosure().Size(); got != 1 {
		t.Fatalf("particleClosure size=%d, want1", got)
	}
}

// Original TBParTest.testParticleClosureExampleConstLevel.
func TestJavaTBParParticleClosureExampleConstLevel(t *testing.T) {
	p := NewLNBool(true)
	tbPar := NewTBPar(0)
	evenP := NewLNEven(p)
	phi := NewLNDisj(evenP, NewLNAll(NewLNNeg(p)))
	tbPar.AddElement(phi)
	particleClosure := tbPar.ParticleClosure()
	if !(particleClosure.Size() == 3) {
		t.Fatal("size original assertion failed")
	}
	if !(phi == particleClosure.ParAt(0).ExprAt(0)) {
		t.Fatal("particle0 phi original assertion failed")
	}
	if !(evenP == particleClosure.ParAt(0).ExprAt(1)) {
		t.Fatal("particle0 evenP original assertion failed")
	}
	if !(p == particleClosure.ParAt(0).ExprAt(2)) {
		t.Fatal("particle0 p original assertion failed")
	}
	if !(phi == particleClosure.ParAt(1).ExprAt(0)) {
		t.Fatal("particle1 phi original assertion failed")
	}
	if !(evenP == particleClosure.ParAt(1).ExprAt(1)) {
		t.Fatal("particle1 evenP original assertion failed")
	}
	nextEvenP := particleClosure.ParAt(1).ExprAt(2)
	if !(nextEvenP.Kind == LiveExprNext) {
		t.Fatal("nextEvenP type original assertion failed")
	}
	if !(evenP == nextEvenP.GetBody()) {
		t.Fatal("nextEvenP body original assertion failed")
	}
	if !(phi == particleClosure.ParAt(2).ExprAt(0)) {
		t.Fatal("particle2 phi original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(1).Kind == LiveExprAll) {
		t.Fatal("particle2 all type original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(1).GetBody().Kind == LiveExprNeg) {
		t.Fatal("particle2 all body type original assertion failed")
	}
	if !(p == particleClosure.ParAt(2).ExprAt(1).GetBody().GetBody()) {
		t.Fatal("particle2 all neg body original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(1).String() == "[]-TRUE") {
		t.Fatal("particle2 all string original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(2).Kind == LiveExprNeg) {
		t.Fatal("particle2 neg type original assertion failed")
	}
	if !(p == particleClosure.ParAt(2).ExprAt(2).GetBody()) {
		t.Fatal("particle2 neg body original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(2).String() == "-TRUE") {
		t.Fatal("particle2 neg string original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(3).Kind == LiveExprNext) {
		t.Fatal("particle2 next type original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(3).GetBody().Kind == LiveExprAll) {
		t.Fatal("particle2 next body type original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(3).GetBody().GetBody().Kind == LiveExprNeg) {
		t.Fatal("particle2 next all body type original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(3).String() == "()[]-TRUE") {
		t.Fatal("particle2 next string original assertion failed")
	}
}

// Original TBParTest.testParticleClosureExampleStateLevel.
func TestJavaTBParParticleClosureExampleStateLevel(t *testing.T) {
	p := javaLivenessDummyState("p")
	tbPar := NewTBPar(0)
	evenP := NewLNEven(p)
	phi := NewLNDisj(evenP, NewLNAll(NewLNNeg(p)))
	tbPar.AddElement(phi)
	particleClosure := tbPar.ParticleClosure()
	if !(particleClosure.Size() == 3) {
		t.Fatal("size original assertion failed")
	}
	if !(phi == particleClosure.ParAt(0).ExprAt(0)) {
		t.Fatal("particle0 phi original assertion failed")
	}
	if !(evenP == particleClosure.ParAt(0).ExprAt(1)) {
		t.Fatal("particle0 evenP original assertion failed")
	}
	if !(p == particleClosure.ParAt(0).ExprAt(2)) {
		t.Fatal("particle0 p original assertion failed")
	}
	if !(phi == particleClosure.ParAt(1).ExprAt(0)) {
		t.Fatal("particle1 phi original assertion failed")
	}
	if !(evenP == particleClosure.ParAt(1).ExprAt(1)) {
		t.Fatal("particle1 evenP original assertion failed")
	}
	nextEvenP := particleClosure.ParAt(1).ExprAt(2)
	if !(nextEvenP.Kind == LiveExprNext) {
		t.Fatal("nextEvenP type original assertion failed")
	}
	if !(evenP == nextEvenP.GetBody()) {
		t.Fatal("nextEvenP body original assertion failed")
	}
	if !(phi == particleClosure.ParAt(2).ExprAt(0)) {
		t.Fatal("particle2 phi original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(1).Kind == LiveExprAll) {
		t.Fatal("particle2 all type original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(1).GetBody().Kind == LiveExprNeg) {
		t.Fatal("particle2 all body type original assertion failed")
	}
	if !(p == particleClosure.ParAt(2).ExprAt(1).GetBody().GetBody()) {
		t.Fatal("particle2 all neg body original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(1).String() == "[]-p") {
		t.Fatal("particle2 all string original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(2).Kind == LiveExprNeg) {
		t.Fatal("particle2 neg type original assertion failed")
	}
	if !(p == particleClosure.ParAt(2).ExprAt(2).GetBody()) {
		t.Fatal("particle2 neg body original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(2).String() == "-p") {
		t.Fatal("particle2 neg string original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(3).Kind == LiveExprNext) {
		t.Fatal("particle2 next type original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(3).GetBody().Kind == LiveExprAll) {
		t.Fatal("particle2 next body type original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(3).GetBody().GetBody().Kind == LiveExprNeg) {
		t.Fatal("particle2 next all body type original assertion failed")
	}
	if !(particleClosure.ParAt(2).ExprAt(3).String() == "()[]-p") {
		t.Fatal("particle2 next string original assertion failed")
	}
}
