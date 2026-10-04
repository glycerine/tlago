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
package tlc

import (
	"slices"
	"testing"
)

// Original TableauNodePtrTableTest.testSetDoneBFSOrder.
func TestJavaTableauNodePtrTableSetDoneBFSOrder(tt *testing.T) {

	// Test behavior of TNPT when state graph/fingerprint graph nodes are added in
	// strict BFS order.

	tbl := NewTableauNodePtrTable(0)

	DONE := int64(4711) // This is a disk location (pointer) in practice.
	fp := uint64(42)    // hash(v) = fp
	t := 1
	u := 2

	// 1) A transition from state u -> v is added into the behavior graph. As part
	// of adding u, TLC records v with some tableau ids.
	tbl.PutElem(fp, t, TableauNodePtrTableUndone)
	if tbl.IsDone(fp) {
		tt.Fatal("original assertion 1 failed")
	}
	tbl.PutElem(fp, u, TableauNodePtrTableUndone)
	if tbl.IsDone(fp) {
		tt.Fatal("original assertion 2 failed")
	}

	// 2) v -> ... is added into the behavior graph.
	tbl.SetDone(fp)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 3 failed")
	}

	// Marking the nodes done has become a no-op.
	tbl.PutElem(fp, u, DONE)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 4 failed")
	}
	tbl.PutElem(fp, t, DONE)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 5 failed")
	}

}

// Original TableauNodePtrTableTest.testSetDoneNoOrder.
func TestJavaTableauNodePtrTableSetDoneNoOrder(tt *testing.T) {

	// Test behavior of TNPT when state graph/fingerprint graph nodes are added in
	// non-BFS order.

	tbl := NewTableauNodePtrTable(0)

	DONE := int64(4711) // This is a disk location (pointer) in practice.
	fp := uint64(42)    // hash(v) = fp
	t := 1
	u := 2

	// 1) A transition from state v -> ... is added into the behavior graph. Because
	// v has not been recorded earlier, adding it into the behavior graph is reduced
	// to calling setDone no GraphNodes are recorded.
	tbl.SetDone(fp)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 1 failed")
	}

	// 2) u -> v is added into the behavior graph.
	tbl.PutElem(fp, t, TableauNodePtrTableUndone)
	if tbl.IsDone(fp) {
		tt.Fatal("original assertion 2 failed")
	}
	tbl.PutElem(fp, u, TableauNodePtrTableUndone)
	if tbl.IsDone(fp) {
		tt.Fatal("original assertion 3 failed")
	}

	// Marking the nodes done has become a no-op.
	tbl.PutElem(fp, u, DONE)
	if tbl.IsDone(fp) {
		tt.Fatal("original assertion 4 failed")
	}
	tbl.PutElem(fp, t, DONE)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 5 failed")
	}

}

// Original TableauNodePtrTableTest.testSetDone.
func TestJavaTableauNodePtrTableSetDone(tt *testing.T) {

	tbl := NewTableauNodePtrTable(0) // init with 0 so that grow is tested

	fingerprint := uint64(1)
	if tbl.IsDone(fingerprint) {
		tt.Fatal("original assertion 1 failed")
	}

	tbl.SetDone(fingerprint)
	if !tbl.IsDone(fingerprint) {
		tt.Fatal("original assertion 2 failed")
	}

	tbl.PutElem(fingerprint, 1, TableauNodePtrTableUndone)
	// This ends up as -2 for the high part of the long and thus tbl isn't
	// done anymore.
	if tbl.IsDone(fingerprint) {
		tt.Fatal("original assertion 3 failed")
	}

	// Add another node BUT with different tableau index. The done state
	// does NOT change.
	tbl.PutElem(fingerprint, 2, 4711)
	if tbl.IsDone(fingerprint) {
		tt.Fatal("original assertion 4 failed")
	}

	tbl.SetDone(fingerprint)
	if !tbl.IsDone(fingerprint) {
		tt.Fatal("original assertion 5 failed")
	}

}

// Original TableauNodePtrTableTest.testSetDone2.
func TestJavaTableauNodePtrTableSetDone2(tt *testing.T) {

	tbl := NewTableauNodePtrTable(0)

	DONE := int64(4711) // This is a disk location (pointer) in practice.
	fp := uint64(42)
	t := 1
	u := 2

	// Mark fp as done.
	tbl.SetDone(fp)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 1 failed")
	}

	// Add the behavior graph node <<fp, t>> to tbl and mark it undone.
	tbl.PutElem(fp, t, TableauNodePtrTableUndone)

	// Adding <<fp, t>> to tbl causes fp to become undone again!!!
	if tbl.IsDone(fp) {
		tt.Fatal("original assertion 2 failed")
	}

	// Add a second node <<fp, u>> to the behavior graph (same fingerprint but different
	// tableau node).
	tbl.PutElem(fp, u, TableauNodePtrTableUndone)

	// Nothing changes WRT fp.
	if tbl.IsDone(fp) {
		tt.Fatal("original assertion 3 failed")
	}

	// Marking the additional node done has no effect on fp's done state.
	tbl.PutElem(fp, u, DONE)
	if tbl.IsDone(fp) {
		tt.Fatal("original assertion 4 failed")
	}

	// Marking the *first* node (insertion order) in tbl done, marks fp done again.
	tbl.PutElem(fp, t, DONE)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 5 failed")
	}

}

// Original TableauNodePtrTableTest.testSetDone3.
func TestJavaTableauNodePtrTableSetDone3(tt *testing.T) {

	tbl := NewTableauNodePtrTable(0)

	DONE := int64(4711) // This is a disk location (pointer) in practice.
	fp := uint64(42)
	t := 1
	u := 2

	tbl.SetDone(fp)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 1 failed")
	}

	// fp becomes undone again by recording a node (see dgragh.recordNode)
	tbl.PutElem(fp, t, TableauNodePtrTableUndone)
	if tbl.IsDone(fp) {
		tt.Fatal("original assertion 2 failed")
	}

	// nothing changes if we record additional nodes.
	tbl.PutElem(fp, u, TableauNodePtrTableUndone)
	if tbl.IsDone(fp) {
		tt.Fatal("original assertion 3 failed")
	}

	// Mark the initial node done.
	tbl.PutElem(fp, t, DONE)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 4 failed")
	}

}

// Original TableauNodePtrTableTest.testIsDoneSPP.
func TestJavaTableauNodePtrTableIsDoneSPP(tt *testing.T) {

	tbl := NewTableauNodePtrTable(0)

	fp := uint64(42)
	t := 1
	u := 2

	tbl.SetDone(fp)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 1 failed")
	}

	tbl.Put(fp, t)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 2 failed")
	}

	tbl.Put(fp, u)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 3 failed")
	}

}

// Original TableauNodePtrTableTest.testIsDonePPS.
func TestJavaTableauNodePtrTableIsDonePPS(tt *testing.T) {

	tbl := NewTableauNodePtrTable(0)

	fp := uint64(42)
	t := 1
	u := 2

	tbl.Put(fp, t)
	if tbl.IsDone(fp) {
		tt.Fatal("original assertion 1 failed")
	}

	tbl.Put(fp, u)
	if tbl.IsDone(fp) {
		tt.Fatal("original assertion 2 failed")
	}

	tbl.SetDone(fp)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 3 failed")
	}

}

// Original TableauNodePtrTableTest.testIsDonePSP.
func TestJavaTableauNodePtrTableIsDonePSP(tt *testing.T) {

	tbl := NewTableauNodePtrTable(0)

	fp := uint64(42)
	t := 1
	u := 2

	tbl.Put(fp, t)
	if tbl.IsDone(fp) {
		tt.Fatal("original assertion 1 failed")
	}

	tbl.SetDone(fp)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 2 failed")
	}

	tbl.Put(fp, u)
	if !tbl.IsDone(fp) {
		tt.Fatal("original assertion 3 failed")
	}

}

// Original TableauNodePtrTableTest.testRedundantMethodYieldSameResult.
func TestJavaTableauNodePtrTableRedundantMethodYieldSameResult(tt *testing.T) {

	tbl := NewTableauNodePtrTable(0) // init with 0 so that grow is tested

	fingerprint := uint64(1)

	if tbl.GetNodesLoc(fingerprint) != -1 {
		tt.Fatal("original assertion 1 failed")
	}

	loc := tbl.SetDone(fingerprint)
	if !tbl.IsDone(fingerprint) {
		tt.Fatal("original assertion 2 failed")
	}

	if !(tbl.GetNodesLoc(fingerprint) != -1) {
		tt.Fatal("original assertion 3 failed")
	}
	if loc != tbl.GetNodesLoc(fingerprint) {
		tt.Fatal("original assertion 4 failed")
	}

	if !slices.Equal(tbl.GetNodesByLoc(loc), tbl.GetNodes(fingerprint)) {
		tt.Fatal("original assertion 5 failed")
	}

	if tbl.GetLoc(fingerprint, 1) != -1 {
		tt.Fatal("original assertion 6 failed")
	}
	tbl.addElem(fingerprint, 1, 2342)
	// Cannot lookup after addElem
	if tbl.GetLoc(fingerprint, 1) != -1 {
		tt.Fatal("original assertion 7 failed")
	}

	//...have to put instead
	tbl.PutElem(fingerprint, 1, 2342)
	if !(tbl.GetLoc(fingerprint, 1) != -1) {
		tt.Fatal("original assertion 8 failed")
	}

}
